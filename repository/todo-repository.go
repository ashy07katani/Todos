package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"todos/models"
)

func GetAllTodos(ctx context.Context, db *sql.DB, offset int, limit int, userId string) ([]*models.GetTodoResponse, error) {
	query := `select id,name,description,status,created_at from todo where user_id = $1 order by created_at desc limit $2 offset $3`
	rows, err := db.QueryContext(ctx, query, userId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var todos []*models.GetTodoResponse
	for rows.Next() {
		var todo models.GetTodoResponse
		err = rows.Scan(&todo.Id, &todo.Name, &todo.Description, &todo.TaskStatus, &todo.CreatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, &todo)
	}
	return todos, nil
}

func GetTodoByID(ctx context.Context, db *sql.DB, id string, user_id string) (*models.GetTodoResponse, error) {
	var todo = new(models.GetTodoResponse)
	query := `select id,name,description,status,created_at from todo where id= $1 and user_id =$2`
	row := db.QueryRowContext(ctx, query, id, user_id)
	err := row.Scan(&todo.Id, &todo.Name, &todo.Description, &todo.TaskStatus, &todo.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return todo, nil
}

func CreateTodo(ctx context.Context, db *sql.DB, todo *models.Todo, user_id string) error {
	query := `insert into todo (name, description, user_id) values ($1, $2, $3)`
	_, err := db.ExecContext(ctx, query, todo.Name, todo.Description, user_id)
	return err
}

func DeleteTodo(ctx context.Context, db *sql.DB, id string, user_id string) error {
	query := `delete from todo where id=$1 and user_id=$2;`
	rows, _ := db.ExecContext(ctx, query, id, user_id)
	rowCount, err := rows.RowsAffected()
	if rowCount == 0 {
		return fmt.Errorf("no rows found for this user with this Id: %s", id)
	}
	return err
}

func UpdateTodo(ctx context.Context, db *sql.DB, params map[string]interface{}, id string, user_id string) error {

	paramNames := []string{}
	paramValues := []interface{}{}

	var i int = 1
	for key, value := range params {
		paramNames = append(paramNames, fmt.Sprintf("%s=$%d", key, i))
		paramValues = append(paramValues, value)
		i++
	}
	j := i + 1
	query := fmt.Sprintf(`update todo set %s where id = $%d and user_id = $%d`, strings.Join(paramNames, ","), i, j)
	paramValues = append(paramValues, id)
	paramValues = append(paramValues, user_id)
	res, err := db.ExecContext(ctx, query, paramValues...)
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no rows found for id: %s", id)
	}
	return err

}

func SearchTodo(ctx context.Context, db *sql.DB, searchParam string, limit int, offset int, user_id string) ([]*models.GetTodoResponse, error) {
	query := `select id, name, description, status, created_at from todo where user_id =$1 to_tsvector('simple', name || ' ' || description) @@ to_tsquery('simple', $2) limit $3 offset $4`
	fmt.Println(query)
	var todos []*models.GetTodoResponse
	rows, err := db.QueryContext(ctx, query, user_id, searchParam, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var todo models.GetTodoResponse
		err = rows.Scan(&todo.Id, &todo.Name, &todo.Description, &todo.TaskStatus, &todo.CreatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, &todo)
	}
	return todos, nil

}

func CreateUser(ctx context.Context, db *sql.DB, user *models.User) error {
	query := `insert into users (username, email, hashpassword) values ($1,$2,$3)`
	_, err := db.ExecContext(ctx, query, user.UserName, user.Email, user.HashedPassword)
	return err

}

func FetchUserWithUserID(ctx context.Context, db *sql.DB, userName string) (*models.User, error) {
	query := `select id, username, email, hashpassword, created_at  from users where username = $1`
	row := db.QueryRowContext(ctx, query, userName)
	user := new(models.User)
	err := row.Scan(&user.Id, &user.UserName, &user.Email, &user.HashedPassword, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("there is no user with this user-name")
		}
		return nil, err
	}
	return user, nil
}

func SaveRefreshToken(ctx context.Context, db *sql.DB, saveRefresh *models.SaveRefresh) error {
	query := `insert into refresh (user_id, token_hash, expires_at) values ($1, $2, $3)`
	_, err := db.ExecContext(ctx, query, saveRefresh.UserId, saveRefresh.TokenHash, saveRefresh.ExpiresAt)
	return err
}

func FetchRefreshToken(ctx context.Context, db *sql.DB, hashedRefresh string) (*models.SaveRefresh, error) {
	query := `select user_id, token_hash, expires_at, revoked from refresh where token_hash =$1`
	refresh := new(models.SaveRefresh)
	row := db.QueryRowContext(ctx, query, hashedRefresh)
	err := row.Scan(&refresh.UserId, &refresh.TokenHash, &refresh.ExpiresAt, &refresh.Revoked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no token found")
		}
		return nil, err
	}
	return refresh, nil
}

func InvalidateRefreshToken(ctx context.Context, db *sql.DB, hashedRefresh string) error {
	query := `update refresh set revoked = true where token_hash =$1`
	_, err := db.ExecContext(ctx, query, hashedRefresh)
	return err
}

func IsEmailExists(ctx context.Context, db *sql.DB, forgot *models.ForgotPasswordRequest) (string, error) {
	userQuery := `select id from users where email = $1`
	row := db.QueryRow(userQuery, forgot.Email)
	userId := ""
	err := row.Scan(&userId)
	if err != nil {
		return userId, err
	}
	return userId, nil
}

func StoreForgotPasswordToken(ctx context.Context, db *sql.DB, request map[string]string) *models.ErrorResponse {
	email, ok := request["email"]
	errResponse := new(models.ErrorResponse)
	if ok {
		checkExistingTokenQuery := `select expires_at, used from forgotpassword where email= $1 order by created_At desc`
		row := db.QueryRowContext(ctx, checkExistingTokenQuery, email)
		var expires_At time.Time
		var used bool
		if err := row.Scan(&expires_At, &used); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				errResponse.Message = err.Error()
				errResponse.Status = http.StatusInternalServerError
				return errResponse
			}
		} else {
			if expires_At.After(time.Now()) && !used {
				errResponse.Message = "an active link has already been sent to your mail, for new token wait for sometime"
				errResponse.Status = http.StatusBadRequest
				return errResponse
			}
		}
		forgotPasswordQuery := `insert into forgotpassword (userid,email,token) values ($1, $2, $3)`
		_, err := db.ExecContext(ctx, forgotPasswordQuery, request["userid"], email, request["token"])
		if err != nil {
			errResponse.Message = err.Error()
			errResponse.Status = http.StatusInternalServerError
			return errResponse
		}
	} else {
		errResponse.Message = "no email received from request"
		errResponse.Status = http.StatusBadRequest
		return errResponse
	}
	return nil
}

func UpdatePassword(ctx context.Context, db *sql.DB, request *models.UpdatePasswordRequest, token string) *models.ErrorResponse {
	transaction, err := db.BeginTx(ctx, nil)
	errResponse := new(models.ErrorResponse)
	if err != nil {
		errResponse.Message = err.Error()
		errResponse.Status = http.StatusInternalServerError
		return errResponse
	}
	defer transaction.Rollback()
	query := `select userid, expires_At, used from forgotpassword where token=$1 for update`
	row := transaction.QueryRowContext(ctx, query, token)
	userId := ""
	expiresAt := time.Now()
	used := false
	if err = row.Scan(&userId, &expiresAt, &used); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errResponse.Message = "no rows with this token"
			errResponse.Status = http.StatusNotFound
			return errResponse
		}
		errResponse.Message = fmt.Sprintf("something went wrong while processing the token : %s", err.Error())
		errResponse.Status = http.StatusInternalServerError
		return errResponse

	}

	if expiresAt.Before(time.Now()) || used {
		errResponse.Message = "provided token has either already been used or is expired"
		errResponse.Status = http.StatusInternalServerError
		return errResponse
	}

	updatePasswordQuery := `update users set hashpassword = $1 where id = $2`
	rows, err := transaction.ExecContext(ctx, updatePasswordQuery, request.NewPassword, userId)
	if err != nil {
		errResponse.Message = "something went wrong while updating the password"
		errResponse.Status = http.StatusInternalServerError
		return errResponse
	}
	rowsAffected, _ := rows.RowsAffected()
	if rowsAffected == 0 {
		errResponse.Message = "no user affected"
		errResponse.Status = http.StatusInternalServerError
		return errResponse
	}

	changeTokenStatusQuery := `update forgotpassword set used = true where token = $1`
	rows, err = transaction.ExecContext(ctx, changeTokenStatusQuery, token)
	if err != nil {
		errResponse.Message = "something went wrong while updating the password"
		errResponse.Status = http.StatusInternalServerError
		return errResponse
	}
	rowsAffected, _ = rows.RowsAffected()
	if rowsAffected == 0 {
		errResponse.Message = "no password affected"
		errResponse.Status = http.StatusInternalServerError
		return errResponse
	}

	err = transaction.Commit()
	if err != nil {
		errResponse.Message = "update password has failed, rolling back"
		errResponse.Status = http.StatusInternalServerError
		return errResponse
	}
	return nil
}
