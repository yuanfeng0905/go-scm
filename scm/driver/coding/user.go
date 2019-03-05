// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coding

import (
	"context"
	"fmt"

	"github.com/drone/go-scm/scm"
)

type userService struct {
	client *wrapper
}

func (s *userService) Find(ctx context.Context) (*scm.User, *scm.Response, error) {
	userRet := new(userRet)
	res, err := s.client.do(ctx, "GET", "api/account/current_user", nil, userRet)
	if err != nil {
		return &scm.User{}, res, err
	}
	// 第二次请求获取用户邮箱
	emailRet := new(emailRet)
	res1, err1 := s.client.do(ctx, "GET", "api/account/email", nil, emailRet)
	if err1 != nil {
		return &scm.User{}, res1, err1
	}

	u := convertUser(&userRet.Data)
	u.Email = emailRet.Data

	return u, res, nil
}

func (s *userService) FindLogin(ctx context.Context, login string) (*scm.User, *scm.Response, error) {
	path := fmt.Sprintf("api/account/key/%s", login)
	out := new(userRet)
	res, err := s.client.do(ctx, "GET", path, nil, out)

	return convertUser(&out.Data), res, err
}

func (s *userService) FindEmail(ctx context.Context) (string, *scm.Response, error) {
	user, res, err := s.Find(ctx)
	return user.Email, res, err
}

//
// native data structures
//
type user struct {
	GlobalKey string `json:"global_key"`
	Fullname  string `json:"name"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
}

type emailRet struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

type userRet struct {
	Code int  `json:"code"`
	Data user `json:"data"`
}

func convertUser(src *user) *scm.User {
	return &scm.User{
		Login:  userLogin(src),
		Avatar: src.Avatar,
		Email:  src.Email,
		Name:   src.Fullname,
	}
}

func userLogin(src *user) string {
	return src.GlobalKey
}
