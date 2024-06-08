package execute

import (
	"HospitalManager/db/scylla/scylladb"
	"HospitalManager/dto/req/user_req"
	"HospitalManager/helper"
	"HospitalManager/model"
	"HospitalManager/security"
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/scylladb/gocqlx/v2/qb"
	"time"
)

func (q *Queries) CreateAdminAccount() error {
	tableName := fmt.Sprintf("%s.doctors", q.keyspace)
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	password := security.HashAndSalt([]byte("son"))
	insert := &model.Users{
		Id:         id,
		DoctorCode: "Admin",
		Password:   password,
		Fullname:   "Admin",
		Email:      "Admin@gmail.com",
		Phone:      "0923151911",
		Role:       "HEAD_DOCTOR",
		CreateAt:   time.Now(),
		UpdateAt:   time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "doctor_code", "password",
			"fullname", "email", "phone", "role", "create_at", "update_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) InsertUser(req user_req.RegisterReq) error {
	tableName := fmt.Sprintf("%s.doctors", q.keyspace)
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	checkDoctorCode, _ := q.GetUserByOption(req.DoctorCode, "doctor_code")
	if len(checkDoctorCode) > 0 {
		return errors.New("DoctorCode duplicate")
	}
	checkEmail, _ := q.GetUserByOption(req.Email, "email")
	if len(checkEmail) > 0 {
		return errors.New("Email duplicate")
	}
	password := security.HashAndSalt([]byte(req.Password))
	insert := &model.Users{
		Id:         id,
		DoctorCode: req.DoctorCode,
		Password:   password,
		Fullname:   req.Fullname,
		Email:      req.Email,
		Phone:      req.Phone,
		Role:       "DOCTOR",
		CreateAt:   time.Now(),
		UpdateAt:   time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "doctor_code", "password",
			"fullname", "email", "phone", "role", "create_at", "update_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) GetUserByOption(value string, option string) ([]model.Users, error) {
	tableName := fmt.Sprintf("%s.doctors", q.keyspace)
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var users []model.Users
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option: value,
	})
	if err := query.SelectRelease(&users); err != nil {
		return []model.Users{}, err
	}
	if len(users) == 0 {
		return []model.Users{}, errors.New("No Doctor Data Found")
	}
	return users, nil
}

func (q *Queries) Validate(req user_req.LoginReq, c echo.Context) (model.Users, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	users, err := q.GetUserByOption(req.DoctorCode, "doctor_code")
	if err != nil {
		return model.Users{}, err
	}
	isValid := security.ComparePassword(users[0].Password, []byte(req.Password))
	if !isValid {
		return model.Users{}, errors.New("Username or Password is incorrect!")
	}
	return users[0], nil
}

func (q *Queries) Logout(c echo.Context) error {
	helper.DeleteCookie(c, "jwt")
	helper.DeleteCookie(c, "refresh-token")
	return nil
}

func (q *Queries) RefreshToken(c echo.Context) (string, error) {
	cookie, err := c.Cookie("refresh-token")
	if err != nil {
		return "", err
	}
	newAccessToken, err := security.GenAccessTokenFromRefreshToken(cookie.Value)
	if err != nil {
		return "", err
	}
	helper.DeleteCookie(c, "jwt")
	helper.CreateCookie(c, "jwt", newAccessToken, 3600)
	return newAccessToken, nil
}

func (q *Queries) GetAllUsers() ([]model.Users, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.doctors", q.keyspace)
	var users []model.Users
	stmt, names := qb.Select(tableName).
		ToCql()
	query := q.session.Query(stmt, names)
	if err := query.SelectRelease(&users); err != nil {
		return []model.Users{}, err
	}
	return users, nil
}

func (q *Queries) UpdateProfile(req *scylladb.UpdateProfileReq, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.doctors", q.keyspace)
	Userid := c.Get("Userid")
	user, _ := q.GetUserByOption(req.Email, "email")
	if len(user) > 0 && user[0].Id.String() != Userid {
		return errors.New("email duplicate 2")
	}
	stmt, names := qb.Update(tableName).
		Set("fullname").
		Set("email").
		Set("phone").
		Set("update_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(req)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) ChangePermission(DoctorCode string, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.doctors", q.keyspace)
	user, err := q.GetUserByOption(DoctorCode, "doctor_code")
	if err != nil {
		return err
	}
	curRole := user[0].Role
	curId := user[0].Id.String()
	if curRole == "DOCTOR" {
		update := scylladb.UpdateRole{
			Role: "HEAD_DOCTOR",
			Id:   curId,
		}
		stmt, names := qb.Update(tableName).
			Set("role").
			Where(qb.Eq("id")).
			ToCql()

		query := q.session.Query(stmt, names).BindStruct(update)
		if err := query.ExecRelease(); err != nil {
			return err
		}

		accessToken, err := security.GenToken(curId, "HEAD_DOCTOR", time.Hour)
		if err != nil {
			return errors.New("Gen Access Token Fail, Try again")
		}
		refreshToken, err := security.GenToken(curId, "HEAD_DOCTOR", time.Hour*24*7)
		if err != nil {
			return errors.New("Gen Refresh Token Fail, Try again")
		}
		helper.CreateCookie(c, "jwt", accessToken, 3600*7*24)
		helper.CreateCookie(c, "refresh-token", refreshToken, 3600*7*24)

		return nil
	}
	update := scylladb.UpdateRole{
		Role: "DOCTOR",
		Id:   curId,
	}
	stmt, names := qb.Update(tableName).
		Set("role").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) GetProfileCurrent(c echo.Context) (model.Users, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var user model.Users
	Userid := c.Get("Userid")
	tableName := fmt.Sprintf("%s.doctors", q.keyspace)
	stmt, names := qb.Select(tableName).
		Where(qb.Eq("id")).
		ToCql()
	query := q.session.Query(stmt, names).BindMap(qb.M{
		"id": Userid,
	})
	if err := query.GetRelease(&user); err != nil {
		return model.Users{}, err
	}
	return user, nil
}

func (q *Queries) ChangePassword(req user_req.ChangePswReq, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.doctors", q.keyspace)
	userId := c.Get("Userid")
	idStr := fmt.Sprintf("%v", userId)
	users, err := q.GetUserByOption(idStr, "id")
	if err != nil {
		return err
	}
	isValid := security.ComparePassword(users[0].Password, []byte(req.OldPassword))
	if !isValid {
		return errors.New("Password is incorrect!")
	}
	password := security.HashAndSalt([]byte(req.NewPassword))
	update := scylladb.ChangePsw{
		Password: password,
		Id:       idStr,
	}
	stmt, names := qb.Update(tableName).
		Set("password").
		Where(qb.Eq("id")).
		ToCql()
	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) InsertResetToken(idDoctor, value string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.reset_token", q.keyspace)
	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	insert := &model.ResetToken{
		Id:       id,
		IdDoctor: idDoctor,
		Value:    value,
	}
	stmt := qb.Insert(tableName).
		Columns("id", "id_doctor", "value").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) GetResetToken(option string, value string) (model.ResetToken, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var token model.ResetToken
	tableName := fmt.Sprintf("%s.reset_token", q.keyspace)
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		AllowFiltering().
		ToCql()
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option: value,
	})
	if err := query.GetRelease(&token); err != nil {
		return model.ResetToken{}, err
	}
	return token, nil
}

func (q *Queries) UpdateToken(id, value string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.reset_token", q.keyspace)
	update := scylladb.UpdateToken{
		Value: value,
		Id:    id,
	}
	stmt, names := qb.Update(tableName).
		Set("value").
		Where(qb.Eq("id")).
		ToCql()
	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) DeleteToken(id string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.reset_token", q.keyspace)
	delete := scylladb.UpdateToken{
		Id: id,
	}
	stmt, names := qb.Delete(tableName).
		Where(qb.Eq("id")).
		ToCql()
	query := q.session.Query(stmt, names).BindStruct(delete)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) ResetPassword(token, newPassword string) error {
	tableName := fmt.Sprintf("%s.doctors", q.keyspace)
	resetToken, err := security.ValidateToken(token)
	if err != nil {
		return err
	}
	if !resetToken.Valid {
		return errors.New("token not available")
	}
	claims := resetToken.Claims.(jwt.MapClaims)
	userid := claims["userid"].(string)
	password := security.HashAndSalt([]byte(newPassword))
	update := scylladb.ChangePsw{
		Password: password,
		Id:       userid,
	}
	stmt, names := qb.Update(tableName).
		Set("password").
		Where(qb.Eq("id")).
		ToCql()
	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}
