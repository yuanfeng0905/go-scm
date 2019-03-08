// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coding

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
)

type repositoryService struct {
	client *wrapper
}

func (s *repositoryService) Find(ctx context.Context, repo string) (*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s", repo)
	out := new(repository)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertRepository(out), res, err
}

func (s *repositoryService) FindHook(ctx context.Context, repo string, id string) (*scm.Hook, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/hooks/%s", repo, id)
	out := new(hook)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertHook(out), res, err
}

func (s *repositoryService) FindPerms(ctx context.Context, repo string) (*scm.Perm, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s", repo)
	out := new(repository)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertRepository(out).Perm, res, err
}

// 获取当前用户的仓库列表
func (s *repositoryService) List(ctx context.Context, opts scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("api/user/projects?%s", encodeListOptions(opts))
	out := &projectRet{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)

	// 非restful标准的接口，自己处理好分页参数
	if out.Data.Page < out.Data.TotalPage {
		res.Page.Next = out.Data.Page + 1
		res.Page.NextURL = "https://coding.net" //兼容
	} else {
		res.Page.Next = 0
		res.Page.NextURL = ""
	}
	return convertRepositoryList(out), res, err
}

func (s *repositoryService) ListHooks(ctx context.Context, repo string, _ scm.ListOptions) ([]*scm.Hook, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/hooks", repo)
	out := []*hook{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertHookList(out), res, err
}

func (s *repositoryService) ListStatus(ctx context.Context, repo string, ref string, _ scm.ListOptions) ([]*scm.Status, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/statuses/%s", repo, ref)
	out := []*status{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertStatusList(out), res, err
}

func (s *repositoryService) CreateHook(ctx context.Context, repo string, input *scm.HookInput) (*scm.Hook, *scm.Response, error) {
	// target, err := url.Parse(input.Target)
	// if err != nil {
	// 	return nil, nil, err
	// }
	// params := target.Query()
	// params.Set("secret", input.Secret)
	// target.RawQuery = params.Encode()

	// 默认repo格式为owner/repo
	sp := strings.Split(repo, "/")
	inParams := url.Values{}
	inParams.Set("hook_url", input.Target)
	inParams.Set("token", input.Secret)
	if input.Events.PullRequest {
		inParams.Set("type_mr_pr", "on")
	} else if input.Events.Push {
		inParams.Set("type_push", "on")
	}

	path := fmt.Sprintf("api/user/%s/project/%s/git/hook?%s", sp[0], sp[1], inParams.Encode())

	out := new(hook)
	res, err := s.client.do(ctx, "POST", path, nil, out)
	return convertHook(out), res, err
}

func (s *repositoryService) CreateStatus(ctx context.Context, repo string, ref string, input *scm.StatusInput) (*scm.Status, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/statuses/%s", repo, ref)
	in := &statusInput{
		State:       convertFromState(input.State),
		Context:     input.Label,
		Description: input.Desc,
		TargetURL:   input.Target,
	}
	out := new(status)
	res, err := s.client.do(ctx, "POST", path, in, out)
	return convertStatus(out), res, err
}

func (s *repositoryService) DeleteHook(ctx context.Context, repo string, id string) (*scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/hooks/%s", repo, id)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}

//
// native data structures
//

type (
	projectData struct {
		List      []*repository `json:"list"`
		Page      int           `json:"page"`
		PageSize  int           `json:"pageSize"`
		TotalPage int           `json:"totalPage"`
		TotalRow  int           `json:"totalRow"`
	}
	projectRet struct {
		Code int         `json:"code"`
		Data projectData `json:"data"`
	}

	// coding repository resource.
	repository struct {
		ID            int    `json:"id"`
		OwnerUsername string `json:"owner_user_name"`
		Owner         struct {
			ID        int    `json:"id"`
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
		} `json:"owner"`
		Name     string `json:"name"`
		FullName string `json:"display_name"`
		//Private       bool      `json:"private"`
		Fork          bool   `json:"forked"`
		HTMLURL       string `json:"https_url"`
		SSHURL        string `json:"ssh_url"`
		CloneURL      string `json:"git_url"`
		DefaultBranch string `json:"default_branch"`
		CreatedAt     int64  `json:"created_at"`
		UpdatedAt     int64  `json:"updated_at"`
		//Permissions   perm      `json:"permissions"`
	}

	// coding permissions details.
	perm struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
		Pull  bool `json:"pull"`
	}

	// coding hook resource.
	hook struct {
		ID     int        `json:"id"`
		Type   string     `json:"type"`
		Events []string   `json:"events"`
		Active bool       `json:"active"`
		Config hookConfig `json:"config"`
	}

	// coding hook configuration details.
	hookConfig struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		Secret      string `json:"secret"`
	}

	// coding status resource.
	status struct {
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		State       string    `json:"status"`
		TargetURL   string    `json:"target_url"`
		Description string    `json:"description"`
		Context     string    `json:"context"`
	}

	// coding status creation request.
	statusInput struct {
		State       string `json:"state"`
		TargetURL   string `json:"target_url"`
		Description string `json:"description"`
		Context     string `json:"context"`
	}
)

//
// native data structure conversion
//

func convertRepositoryList(src *projectRet) []*scm.Repository {
	var dst []*scm.Repository
	for _, v := range src.Data.List {
		dst = append(dst, convertRepository(v))
	}
	return dst
}

func convertRepository(src *repository) *scm.Repository {
	r := &scm.Repository{
		ID:        strconv.Itoa(src.ID),
		Namespace: src.Owner.Login,
		Name:      src.Name,
		Perm:      convertPerm(perm{}),
		SCM:       "coding",
		Branch:    src.DefaultBranch,
		Private:   true, //默认项目私有
		Clone:     src.CloneURL,
		CloneSSH:  src.SSHURL,
		Created:   time.Unix(src.CreatedAt/1000, 0),
		Updated:   time.Unix(src.UpdatedAt/1000, 0),
	}

	// hook
	// 获取项目列表时，namespace从owner_user_name取
	if src.OwnerUsername != "" && r.Namespace == "" {
		r.Namespace = src.OwnerUsername
	}
	return r
}

// 默认所有coding项目给最高权限
func convertPerm(src perm) *scm.Perm {
	return &scm.Perm{
		Push:  true,
		Pull:  true,
		Admin: true,
	}
}

func convertHookList(src []*hook) []*scm.Hook {
	var dst []*scm.Hook
	for _, v := range src {
		dst = append(dst, convertHook(v))
	}
	return dst
}

func convertHook(from *hook) *scm.Hook {
	return &scm.Hook{
		ID:     strconv.Itoa(from.ID),
		Active: from.Active,
		Target: from.Config.URL,
		Events: from.Events,
	}
}

func convertHookEvent(from scm.HookEvents) []string {
	var events []string
	if from.PullRequest {
		events = append(events, "pull_request")
	}
	if from.Issue {
		events = append(events, "issues")
	}
	if from.IssueComment || from.PullRequestComment {
		events = append(events, "issue_comment")
	}
	if from.Branch || from.Tag {
		events = append(events, "create")
		events = append(events, "delete")
	}
	if from.Push {
		events = append(events, "push")
	}
	return events
}

func convertStatusList(src []*status) []*scm.Status {
	var dst []*scm.Status
	for _, v := range src {
		dst = append(dst, convertStatus(v))
	}
	return dst
}

func convertStatus(from *status) *scm.Status {
	return &scm.Status{
		State:  convertState(from.State),
		Label:  from.Context,
		Desc:   from.Description,
		Target: from.TargetURL,
	}
}

func convertState(from string) scm.State {
	switch from {
	case "error":
		return scm.StateError
	case "failure":
		return scm.StateFailure
	case "pending":
		return scm.StatePending
	case "success":
		return scm.StateSuccess
	default:
		return scm.StateUnknown
	}
}

func convertFromState(from scm.State) string {
	switch from {
	case scm.StatePending, scm.StateRunning:
		return "pending"
	case scm.StateSuccess:
		return "success"
	case scm.StateFailure:
		return "failure"
	default:
		return "error"
	}
}
