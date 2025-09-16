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
