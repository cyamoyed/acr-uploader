package acr

import (
	"encoding/json"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/sirupsen/logrus"
)

type Client struct {
	client    *cr.Client
	regionId  string
}

type Namespace struct {
	Name        string
	RegionId    string
	Status      string
	CreateTime  string
}

type Repository struct {
	Name        string
	Namespace   string
	RepoType    string
	Status      string
	CreateTime  string
	UpdateTime  string
}

type namespaceListResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Namespaces []struct {
			Namespace       string `json:"namespace"`
			NamespaceStatus string `json:"namespaceStatus"`
			AuthorizeType   string `json:"authorizeType"`
		} `json:"namespaces"`
	} `json:"data"`
}

type repoListResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Repos []struct {
			RepoName      string  `json:"repoName"`
			RepoNamespace string  `json:"repoNamespace"`
			RepoType      string  `json:"repoType"`
			RepoStatus    string  `json:"repoStatus"`
			GmtCreate     float64 `json:"gmtCreate"`
			GmtModified   float64 `json:"gmtModified"`
		} `json:"repos"`
	} `json:"data"`
}

type createRepoResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Repo struct {
			RepoName      string `json:"repoName"`
			NamespaceName string `json:"namespaceName"`
			RepoType      string `json:"repoType"`
			Status        string `json:"status"`
			CreateTime    string `json:"createTime"`
			UpdateTime    string `json:"updateTime"`
		} `json:"repo"`
	} `json:"data"`
}

func NewClient(regionId, accessKeyId, accessKeySecret string) (*Client, error) {
	client, err := cr.NewClientWithAccessKey(regionId, accessKeyId, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建ACR客户端失败: %w", err)
	}

	return &Client{client: client, regionId: regionId}, nil
}

func (c *Client) ListNamespaces() ([]Namespace, error) {
	request := cr.CreateGetNamespaceListRequest()
	request.Scheme = "https"

	response, err := c.client.GetNamespaceList(request)
	if err != nil {
		return nil, fmt.Errorf("获取命名空间列表失败: %w", err)
	}

	httpContent := response.GetHttpContentString()
	logrus.Debugf("命名空间API响应(HTTP状态码: %d): %s", response.GetHttpStatus(), truncateString(httpContent, 2000))
	
	if httpContent == "" {
		return nil, fmt.Errorf("API响应为空，请检查网络连接或阿里云配置")
	}

	var nsList namespaceListResponse
	if err := json.Unmarshal([]byte(httpContent), &nsList); err != nil {
		return nil, fmt.Errorf("解析命名空间列表响应失败: %w, 响应内容: %s", err, truncateString(httpContent, 500))
	}

	if nsList.Code != "" && nsList.Code != "200" {
		return nil, fmt.Errorf("API调用失败: %s - %s", nsList.Code, nsList.Message)
	}

	if len(nsList.Data.Namespaces) == 0 {
		return nil, fmt.Errorf("当前账号下没有可用的命名空间，请先在阿里云控制台创建")
	}

	namespaces := make([]Namespace, 0, len(nsList.Data.Namespaces))
	for _, ns := range nsList.Data.Namespaces {
		namespaces = append(namespaces, Namespace{
			Name:     ns.Namespace,
			Status:   ns.NamespaceStatus,
			RegionId: c.regionId,
		})
	}

	logrus.Debugf("成功获取到 %d 个命名空间", len(namespaces))
	return namespaces, nil
}

func (c *Client) ListRepositories(namespace string) ([]Repository, error) {
	request := cr.CreateGetRepoListByNamespaceRequest()
	request.Scheme = "https"
	request.RepoNamespace = namespace
	request.PageSize = requests.Integer("100")
	request.Page = requests.Integer("1")

	response, err := c.client.GetRepoListByNamespace(request)
	if err != nil {
		return nil, fmt.Errorf("获取镜像仓库列表失败: %w", err)
	}

	httpContent := response.GetHttpContentString()
	logrus.Debugf("镜像仓库API响应(HTTP状态码: %d, 命名空间: %s): %s", response.GetHttpStatus(), namespace, truncateString(httpContent, 2000))
	
	if httpContent == "" {
		return nil, fmt.Errorf("API响应为空")
	}

	var repoList repoListResponse
	if err := json.Unmarshal([]byte(httpContent), &repoList); err != nil {
		logrus.Warnf("解析镜像仓库列表响应失败: %v, 原始响应: %s", err, truncateString(httpContent, 1000))
		
		var rawResp map[string]interface{}
		if err := json.Unmarshal([]byte(httpContent), &rawResp); err == nil {
			logrus.Warnf("解析后的原始响应结构: %v", rawResp)
		}
		
		return nil, fmt.Errorf("解析镜像仓库列表响应失败: %w", err)
	}

	if repoList.Code != "" && repoList.Code != "200" {
		return nil, fmt.Errorf("API调用失败: %s - %s", repoList.Code, repoList.Message)
	}

	if len(repoList.Data.Repos) == 0 {
		logrus.Warnf("当前命名空间 [%s] 下没有镜像仓库", namespace)
		
		var rawResp map[string]interface{}
		if err := json.Unmarshal([]byte(httpContent), &rawResp); err == nil {
			logrus.Warnf("响应数据结构: %v", rawResp)
		}
	}

	repositories := make([]Repository, 0, len(repoList.Data.Repos))
	for _, repo := range repoList.Data.Repos {
		repositories = append(repositories, Repository{
			Name:       repo.RepoName,
			Namespace:  repo.RepoNamespace,
			RepoType:   repo.RepoType,
			Status:     repo.RepoStatus,
			CreateTime: fmt.Sprintf("%.0f", repo.GmtCreate),
			UpdateTime: fmt.Sprintf("%.0f", repo.GmtModified),
		})
	}

	logrus.Debugf("成功获取到 %d 个镜像仓库", len(repositories))
	return repositories, nil
}

func (c *Client) CreateRepository(namespace, repoName, repoType string) (*Repository, error) {
	request := cr.CreateCreateRepoRequest()
	request.Scheme = "https"
	
	request.PathParams = map[string]string{
		"RepoNamespace": namespace,
		"RepoName":      repoName,
	}
	
	request.Headers = map[string]string{
		"Content-Type": "application/json",
	}
	
	request.SetContent([]byte(fmt.Sprintf(`{"repoType": "%s"}`, repoType)))

	response, err := c.client.CreateRepo(request)
	if err != nil {
		return nil, fmt.Errorf("创建镜像仓库失败: %w", err)
	}

	httpContent := response.GetHttpContentString()
	logrus.Debugf("创建仓库API响应: %s", truncateString(httpContent, 2000))

	var createResp createRepoResponse
	if err := json.Unmarshal([]byte(httpContent), &createResp); err != nil {
		return nil, fmt.Errorf("解析创建仓库响应失败: %w, 响应内容: %s", err, truncateString(httpContent, 500))
	}

	if createResp.Code != "" && createResp.Code != "200" {
		return nil, fmt.Errorf("API调用失败: %s - %s", createResp.Code, createResp.Message)
	}

	repo := createResp.Data.Repo
	return &Repository{
		Name:       repo.RepoName,
		Namespace:  repo.NamespaceName,
		RepoType:   repo.RepoType,
		Status:     repo.Status,
		CreateTime: repo.CreateTime,
		UpdateTime: repo.UpdateTime,
	}, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (c *Client) TestConnection() error {
	_, err := c.ListNamespaces()
	return err
}