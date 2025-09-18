package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
	//$1
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

/*
 user_id
ashish-tripathi(# token_hash varchar(1000),
ashish-tripathi(# expires_at timestamp,
ashish-tripathi(# revoked boolean default false)

*/
