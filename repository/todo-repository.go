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

func GetAllTodos(ctx context.Context, db *sql.DB, offset int, limit int) ([]*models.Todo, error) {
	query := `select id,name,description,status,created_at from todo order by created_at desc limit $1 offset $2`
	rows, err := db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var todos []*models.Todo
	for rows.Next() {
		var todo models.Todo
		err = rows.Scan(&todo.Id, &todo.Name, &todo.Description, &todo.TaskStatus, &todo.CreatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, &todo)
	}
	return todos, nil
}

func GetTodoByID(ctx context.Context, db *sql.DB, id string) (*models.Todo, error) {
	var todo = new(models.Todo)
	query := `select id,name,description,status,created_at from todo where id= $1`
	row := db.QueryRowContext(ctx, query, id)
	err := row.Scan(&todo.Id, &todo.Name, &todo.Description, &todo.TaskStatus, &todo.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return todo, nil
}

func CreateTodo(ctx context.Context, db *sql.DB, todo *models.Todo) error {
	query := `insert into todo (id, name, description) values ($1, $2, $3)`
	_, err := db.ExecContext(ctx, query, todo.Id, todo.Name, todo.Description)
	return err
}

func DeleteTodo(ctx context.Context, db *sql.DB, id string) error {
	query := `delete from todo where id=$1;`
	_, err := db.ExecContext(ctx, query, id)
	return err
}

func UpdateTodo(ctx context.Context, db *sql.DB, params map[string]interface{}, id string) error {

	paramNames := []string{}
	paramValues := []interface{}{}

	var i int = 1
	for key, value := range params {
		paramNames = append(paramNames, fmt.Sprintf("%s=$%d", key, i))
		paramValues = append(paramValues, value)
		i++
	}
	query := fmt.Sprintf(`update todo set %s where id = $%d`, strings.Join(paramNames, ","), i)
	paramValues = append(paramValues, id)
	res, err := db.ExecContext(ctx, query, paramValues...)
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no rows found for id: %s", id)
	}
	return err

}

func SearchTodo(ctx context.Context, db *sql.DB, searchParam string, limit int, offset int) ([]*models.Todo, error) {
	query := `select id, name, description, status, created_at from todo where to_tsvector('simple', name || ' ' || description) @@ to_tsquery('simple', $1) limit $2 offset $3`
	fmt.Println(query)
	var todos []*models.Todo
	rows, err := db.QueryContext(ctx, query, searchParam, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var todo models.Todo
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
