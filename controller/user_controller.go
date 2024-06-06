package controller

import (
	"HospitalManager/db/scylla/scylladb"
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/dto/req/user_req"
	"HospitalManager/dto/res"
	"HospitalManager/security"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type UserController struct {
	Queries *execute.Queries
}

func (u *UserController) CreateAdminAccount() error {
	user, err := u.Queries.GetUserByOption("Admin", "doctor_code")
	if len(user) > 0 {
		return nil
	}
	err = u.Queries.CreateAdminAccount()
	if err != nil {
		return err
	}
	return nil
}

func (u *UserController) Register(c echo.Context) error {
	req := user_req.RegisterReq{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := u.Queries.InsertUser(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "register success!")
}

func (u *UserController) RegisterList(c echo.Context) error {
	var reqs []user_req.RegisterReq
	if err := c.Bind(&reqs); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	var bugs []user_req.RegisterReq

	for _, req := range reqs {
		err := u.Queries.InsertUser(req)
		if err != nil {
			bugs = append(bugs, req)
		}
	}
	if len(bugs) > 0 {
		return c.JSON(http.StatusOK, res.Response{
			Message: "Cannot register users",
			Data:    bugs,
		})
	}
	return c.JSON(http.StatusOK, "register success!")
}

func (u *UserController) Login(c echo.Context) error {
	req := user_req.LoginReq{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	user, err := u.Queries.Validate(req, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	accessToken, err := security.GenToken(user.Id.String(), user.Role, time.Hour)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.New("access token gen fail"))
	}
	refreshToken, err := security.GenToken(user.Id.String(), user.Role, time.Hour*24*7)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.New("refresh token gen fail"))
	}
	return c.JSON(http.StatusOK, res.LoginRes{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (u *UserController) Logout(c echo.Context) error {
	err := u.Queries.Logout(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "log out success")
}

func (u *UserController) Test(c echo.Context) error {
	return c.JSON(http.StatusOK, "Test")
}

func (u *UserController) RefreshToken(c echo.Context) error {
	newAccessToken, err := u.Queries.RefreshToken(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, res.Response{
		Message: "refresh token success",
		Data:    newAccessToken,
	})
}

func (u *UserController) GetAllUsers(c echo.Context) error {
	users, err := u.Queries.GetAllUsers()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, users)
}

func (u *UserController) UpdateProfile(c echo.Context) error {
	userId := c.Get("Userid")
	idStr := fmt.Sprintf("%v", userId)
	req := user_req.UpdateProfileReq{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	update := &scylladb.UpdateProfileReq{
		Fullname: req.Fullname,
		Email:    req.Email,
		Phone:    req.Phone,
		Id:       idStr,
		UpdateAt: time.Now(),
	}
	err := u.Queries.UpdateProfile(update, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "update success")
}

func (u *UserController) ChangePermission(c echo.Context) error {
	doctorCode := c.QueryParam("doctorcode")
	err := u.Queries.ChangePermission(doctorCode, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "change permission success")
}

func (u *UserController) GetProfileCurrent(c echo.Context) error {
	user, err := u.Queries.GetProfileCurrent(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func (u *UserController) ChangePassword(c echo.Context) error {
	req := user_req.ChangePswReq{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := u.Queries.ChangePassword(req, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "change password success")
}
