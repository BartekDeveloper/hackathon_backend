package main

import (
	"bytes"
	_ "database/sql"
	"encoding/csv"
	s "hack/backend/server"
	u "hack/backend/utils"

	"fmt"
	"net/http"
	"reflect" // Added reflect
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

var db s.Database

// goTypeToSQL maps Go types to their corresponding SQL types.
func goTypeToSQL(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "TEXT"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	}
	// Special case for time.Time
	if t.String() == "time.Time" {
		return "TIMESTAMP"
	}
	return "TEXT" // Default to TEXT for other types
}

// GenerateCreateTableSQL generates a CREATE TABLE SQL statement from a Go struct.
func GenerateCreateTableSQL(s interface{}) (string, error) {
	t := reflect.TypeOf(s)
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected a struct")
	}

	tableName := strings.ToLower(t.Name())
	var columns []string
	var primaryKeys []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			continue // Skip fields without a db tag
		}

		columnName := dbTag
		sqlType := goTypeToSQL(field.Type)
		columns = append(columns, fmt.Sprintf(`"%s" %s`, columnName, sqlType))

		if pkTag := field.Tag.Get("pk"); pkTag == "true" {
			primaryKeys = append(primaryKeys, fmt.Sprintf(`"%s"`, columnName))
		}
	}

	if len(columns) == 0 {
		return "", fmt.Errorf("no fields with db tags found in struct")
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (%s`, tableName, strings.Join(columns, ", "))
	if len(primaryKeys) > 0 {
		query += fmt.Sprintf(", PRIMARY KEY (%s)", strings.Join(primaryKeys, ", "))
	}
	query += ");"

	return query, nil
}

func createSchema(db *s.Database) error {
	schemas := []interface{}{
		User{},
		Session{},
		Account{},
		Verification{},
	}

	for _, schema := range schemas {
		query, err := GenerateCreateTableSQL(schema) // Use the local function
		if err != nil {
			return err
		}
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func Router() *gin.Engine {

	router := gin.Default()

	router.LoadHTMLGlob("templates/**/*.html")

	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length")
		c.Next()
	})

	api := router.Group("/")

	router.Use(static.Serve("/assets", static.LocalFile("./public/", true)))

	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		api.POST("/count", func(c *gin.Context) {
			var requestBody CountRequestBody
			if err := c.BindJSON(&requestBody); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong call to API"})
				return
			}

			db.Connect()

			var whereParts []string
			var args []interface{}
			for i, clause := range requestBody.Where {
				if i > 0 {
					whereParts = append(whereParts, clause.Connector)
				}
				sqlOperator := "="
				if clause.Operator != "eq" {
				}
				whereParts = append(whereParts, fmt.Sprintf("\"%s\" %s $%d", clause.Field, sqlOperator, i+1))
				args = append(args, clause.Value)
			}

			whereClause := strings.Join(whereParts, " ")
			sqlQuery := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\" WHERE %s", requestBody.Model, whereClause)

			rows, err := db.Query(sqlQuery, args...)
			if err != nil {
				u.ErrorF("Query Execution Failed:\t%s\n", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Query Execution Failed"})
				return
			}
			defer rows.Close()

			var count int
			if rows.Next() {
				if err := rows.Scan(&count); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
					return
				}
			}

			c.JSON(http.StatusOK, gin.H{"count": count})
		})
		api.POST("/create", func(c *gin.Context) {
			var requestBody CreateRequestBody
			if err := c.BindJSON(&requestBody); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong call to API"})
				return
			}

			db.Connect()

			data := requestBody.Data.(map[string]interface{})
			var columns []string
			var values []interface{}
			var valuePlaceholders []string
			i := 1
			for col, val := range data {
				columns = append(columns, col)
				values = append(values, val)
				valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("$%d", i))
				i++
			}

			quotedColumns := make([]string, len(columns))
			for i, col := range columns {
				quotedColumns[i] = fmt.Sprintf("\"%s\"", col)
			}
			columnNames := strings.Join(quotedColumns, ", ")
			sqlQuery := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES (%s)", requestBody.Model, columnNames, strings.Join(valuePlaceholders, ", "))

			rows, err := db.Query(sqlQuery, values...)
			if err != nil {
				u.ErrorF("Query Execution Failed:\t%s\n", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Query Execution Failed"})
				return
			}
			defer rows.Close()

			c.JSON(http.StatusOK, data)
		})
		api.POST("/delete", func(c *gin.Context) {
			var requestBody DeleteRequestBody
			if err := c.BindJSON(&requestBody); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong call to API"})
				return
			}

			db.Connect()

			var whereParts []string
			var args []interface{}
			for i, clause := range requestBody.Where {
				if i > 0 {
					whereParts = append(whereParts, clause.Connector)
				}
				sqlOperator := "="
				if clause.Operator != "eq" {
				}
				whereParts = append(whereParts, fmt.Sprintf("\"%s\" %s $%d", clause.Field, sqlOperator, i+1))
				args = append(args, clause.Value)
			}

			whereClause := strings.Join(whereParts, " ")
			sqlQuery := fmt.Sprintf("DELETE FROM \"%s\" WHERE %s", requestBody.Model, whereClause)

			rows, err := db.Query(sqlQuery, args...)
			if err != nil {
				u.ErrorF("Query Execution Failed:\t%s\n", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Query Execution Failed"})
				return
			}
			defer rows.Close()

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		api.POST("/delete-many", func(c *gin.Context) {
			var requestBody DeleteManyRequestBody
			if err := c.BindJSON(&requestBody); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong call to API"})
				return
			}

			db.Connect()

			var whereParts []string
			var args []interface{}
			for i, clause := range requestBody.Where {
				if i > 0 {
					whereParts = append(whereParts, clause.Connector)
				}
				sqlOperator := "="
				if clause.Operator != "eq" {
				}
				whereParts = append(whereParts, fmt.Sprintf("\"%s\" %s $%d", clause.Field, sqlOperator, i+1))
				args = append(args, clause.Value)
			}

			whereClause := strings.Join(whereParts, " ")
			sqlQuery := fmt.Sprintf("DELETE FROM \"%s\" WHERE %s", requestBody.Model, whereClause)

			rows, err := db.Query(sqlQuery, args...)
			if err != nil {
				u.ErrorF("Query Execution Failed:\t%s\n", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Query Execution Failed"})
				return
			}
			defer rows.Close()

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		api.POST("/find-many", func(c *gin.Context) {
			var requestBody FindManyRequestBody
			if err := c.BindJSON(&requestBody); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong call to API"})
				return
			}

			db.Connect()

			var whereParts []string
			var args []interface{}
			for i, clause := range requestBody.Where {
				if i > 0 {
					whereParts = append(whereParts, clause.Connector)
				}
				sqlOperator := "="
				if clause.Operator != "eq" {
				}
				whereParts = append(whereParts, fmt.Sprintf("\"%s\" %s $%d", clause.Field, sqlOperator, i+1))
				args = append(args, clause.Value)
			}

			whereClause := strings.Join(whereParts, " ")
			sqlQuery := fmt.Sprintf("SELECT * FROM \"%s\" WHERE %s", requestBody.Model, whereClause)

			rows, err := db.Query(sqlQuery, args...)
			if err != nil {
				u.ErrorF("Query Execution Failed:\t%s\n", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Query Execution Failed"})
				return
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get columns"})
				return
			}

			var results []map[string]interface{}
			for rows.Next() {
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range columns {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
					return
				}

				result := make(map[string]interface{})
				for i, col := range columns {
					val := values[i]
					b, ok := val.([]byte)
					if ok {
						result[col] = string(b)
					} else {
						result[col] = val
					}
				}
				results = append(results, result)
			}

			c.JSON(http.StatusOK, results)
		})
		api.POST("/find-one", func(c *gin.Context) {
			defer func() {
				if r := recover(); r != nil {
					u.ErrorF("Recovered from panic in /find-one: %v", r)
				}
			}()

			var requestBody FindOneRequestBody
			if err := c.BindJSON(&requestBody); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong call to API"})
				return
			}

			if len(requestBody.Where) == 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "A 'where' clause is required for find-one"})
				return
			}

			db.Connect()

			var whereParts []string
			var args []interface{}
			for i, clause := range requestBody.Where {
				if i > 0 {
					whereParts = append(whereParts, clause.Connector)
				}
				sqlOperator := "="
				if clause.Operator != "eq" {
				}
				whereParts = append(whereParts, fmt.Sprintf("\"%s\" %s $%d", clause.Field, sqlOperator, i+1))
				args = append(args, clause.Value)
			}

			whereClause := strings.Join(whereParts, " ")
			sqlQuery := fmt.Sprintf("SELECT * FROM \"%s\" WHERE %s", requestBody.Model, whereClause)

			rows, err := db.Query(sqlQuery, args...)
			if err != nil {
				u.ErrorF("Query: %s, Args: %v", sqlQuery, args)
				u.ErrorF("Query Execution Failed:\t%s\n", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Query Execution Failed from find-one handler"})
				return
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get columns"})
				return
			}

			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			if rows.Next() {
				if err := rows.Scan(valuePtrs...); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
					return
				}
			} else {
				c.JSON(http.StatusOK, gin.H{"error": "empty"})
				return
			}

			result := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				b, ok := val.([]byte)
				if ok {
					result[col] = string(b)
				} else {
					result[col] = val
				}
			}

			c.JSON(http.StatusOK, result)
		})
		api.POST("/update", func(c *gin.Context) {
			var requestBody UpdateRequestBody
			if err := c.BindJSON(&requestBody); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong call to API"})
				return
			}

			db.Connect()

			updateData := requestBody.Update.(map[string]interface{})
			var setParts []string
			var args []interface{}
			i := 1
			for col, val := range updateData {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", col, i))
				args = append(args, val)
				i++
			}

			var whereParts []string
			for _, clause := range requestBody.Where {
				if len(whereParts) > 0 {
					whereParts = append(whereParts, clause.Connector)
				}
				sqlOperator := "="
				if clause.Operator != "eq" {
				}
				whereParts = append(whereParts, fmt.Sprintf("\"%s\" %s $%d", clause.Field, sqlOperator, i))
				args = append(args, clause.Value)
				i++
			}

			setClause := strings.Join(setParts, ", ")
			whereClause := strings.Join(whereParts, " ")
			sqlQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s", requestBody.Model, setClause, whereClause)

			rows, err := db.Query(sqlQuery, args...)
			if err != nil {
				u.ErrorF("Query Execution Failed:\t%s\n", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Query Execution Failed"})
				return
			}
			defer rows.Close()

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		api.POST("/update-many", func(c *gin.Context) {
			var requestBody UpdateManyRequestBody
			if err := c.BindJSON(&requestBody); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong call to API"})
				return
			}

			db.Connect()

			updateData := requestBody.Update.(map[string]interface{})
			var setParts []string
			var args []interface{}
			i := 1
			for col, val := range updateData {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", col, i))
				args = append(args, val)
				i++
			}

			var whereParts []string
			for _, clause := range requestBody.Where {
				if len(whereParts) > 0 {
					whereParts = append(whereParts, clause.Connector)
				}
				sqlOperator := "="
				if clause.Operator != "eq" {
				}
				whereParts = append(whereParts, fmt.Sprintf("\"%s\" %s $%d", clause.Field, sqlOperator, i))
				args = append(args, clause.Value)
				i++
			}

			setClause := strings.Join(setParts, ", ")
			whereClause := strings.Join(whereParts, " ")
			sqlQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s", requestBody.Model, setClause, whereClause)

			rows, err := db.Query(sqlQuery, args...)
			if err != nil {
				u.ErrorF("Query Execution Failed:\t%s\n", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Query Execution Failed"})
				return
			}
			defer rows.Close()

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		api.POST("/create-schema", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Not Implemented"})
		})
	}

	admin := router.Group("/admin")
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"Title":       "Admin Dashboard",
				"Host":        db.Config.Host,
				"Port":        db.Config.Port,
				"User":        db.Config.User,
				"Password":    db.Config.Password,
				"Database":    db.Config.Database,
				"SSL":         db.Config.SSL,
				"IsConnected": db.IsConnected,
			})
		})

		admin.POST("/dashboard/save", func(c *gin.Context) {
			db.Config.Host = c.PostForm("host")
			port, _ := strconv.ParseUint(c.PostForm("port"), 10, 16)
			db.Config.Port = uint16(port)
			db.Config.User = c.PostForm("user")
			db.Config.Password = c.PostForm("password")
			db.Config.Database = c.PostForm("database")
			db.Config.SSL = c.PostForm("ssl")

			// Invalidate current connection
			db.Close()

			// Redirect back to dashboard
			c.Redirect(http.StatusFound, "/admin/dashboard")
		})

		admin.POST("/dashboard/status", func(c *gin.Context) {
			db.Connect()
			c.Redirect(http.StatusFound, "/admin/dashboard")
		})

		admin.POST("/dashboard/export", func(c *gin.Context) {
			format := c.PostForm("format")
			separator := c.PostForm("separator")

			if format == "csv" {
				db.Connect()
				tablesQuery, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed to get tables: %s", err.Error())
					return
				}
				defer tablesQuery.Close()

				var buffer bytes.Buffer
				csvWriter := csv.NewWriter(&buffer)
				csvWriter.Comma = []rune(separator)[0]

				for tablesQuery.Next() {
					var tableName string
					if err := tablesQuery.Scan(&tableName); err != nil {
						c.String(http.StatusInternalServerError, "Failed to scan table name: %s", err.Error())
						return
					}

					rows, err := db.Query(fmt.Sprintf("SELECT * FROM \"%s\"", tableName))
					if err != nil {
						c.String(http.StatusInternalServerError, "Failed to get data from table %s: %s", tableName, err.Error())
						return
					}
					defer rows.Close()

					columns, err := rows.Columns()
					if err != nil {
						c.String(http.StatusInternalServerError, "Failed to get columns from table %s: %s", tableName, err.Error())
						return
					}

					csvWriter.Write(columns)

					for rows.Next() {
						values := make([]interface{}, len(columns))
						valuePtrs := make([]interface{}, len(columns))
						for i := range columns {
							valuePtrs[i] = &values[i]
						}

						if err := rows.Scan(valuePtrs...); err != nil {
							c.String(http.StatusInternalServerError, "Failed to scan row: %s", err.Error())
							return
						}

						var record []string
						for _, val := range values {
							if b, ok := val.([]byte); ok {
								record = append(record, string(b))
							} else {
								record = append(record, fmt.Sprintf("%v", val))
							}
						}
						csvWriter.Write(record)
					}
				}

				csvWriter.Flush()

				c.Header("Content-Description", "File Transfer")
				c.Header("Content-Disposition", "attachment; filename=export.csv")
				c.Data(http.StatusOK, "text/csv", buffer.Bytes())

			} else if format == "sql" {
				db.Connect()
				tablesQuery, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed to get tables: %s", err.Error())
					return
				}
				defer tablesQuery.Close()

				var buffer bytes.Buffer

				for tablesQuery.Next() {
					var tableName string
					if err := tablesQuery.Scan(&tableName); err != nil {
						c.String(http.StatusInternalServerError, "Failed to scan table name: %s", err.Error())
						return
					}

					rows, err := db.Query(fmt.Sprintf("SELECT * FROM \"%s\"", tableName))
					if err != nil {
						c.String(http.StatusInternalServerError, "Failed to get data from table %s: %s", tableName, err.Error())
						return
					}
					defer rows.Close()

					columns, err := rows.Columns()
					if err != nil {
						c.String(http.StatusInternalServerError, "Failed to get columns from table %s: %s", tableName, err.Error())
						return
					}

					for rows.Next() {
						values := make([]interface{}, len(columns))
						valuePtrs := make([]interface{}, len(columns))
						for i := range columns {
							valuePtrs[i] = &values[i]
						}

						if err := rows.Scan(valuePtrs...); err != nil {
							c.String(http.StatusInternalServerError, "Failed to scan row: %s", err.Error())
							return
						}

						var valueStrings []string
						for _, val := range values {
							if b, ok := val.([]byte); ok {
								valueStrings = append(valueStrings, fmt.Sprintf("'%s'", string(b)))
							} else {
								valueStrings = append(valueStrings, fmt.Sprintf("'%v'", val))
							}
						}
						buffer.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n", tableName, strings.Join(columns, ", "), strings.Join(valueStrings, ", ")))
					}
				}

				c.Header("Content-Description", "File Transfer")
				c.Header("Content-Disposition", "attachment; filename=export.sql")
				c.Data(http.StatusOK, "application/sql", buffer.Bytes())

			} else {
				c.String(http.StatusBadRequest, "Invalid export format")
			}
		})

		admin.POST("/dashboard/export-schema", func(c *gin.Context) {
			schemas := []interface{}{
				User{},
				Session{},
				Account{},
				Verification{},
			}

			var buffer bytes.Buffer
			for _, schema := range schemas {
				query, err := GenerateCreateTableSQL(schema) // Use the local function
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed to generate schema: %s", err.Error())
					return
				}
				buffer.WriteString(query)
				buffer.WriteString("\n\n")
			}

			c.Header("Content-Description", "File Transfer")
			c.Header("Content-Disposition", "attachment; filename=schema.sql")
			c.Data(http.StatusOK, "application/sql", buffer.Bytes())
		})

		admin.POST("/dashboard/create-schema", func(c *gin.Context) {
			db.Connect()
			err := createSchema(&db)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to create schema: %s", err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/dashboard")
		})

		admin.POST("/dashboard/get-data", func(c *gin.Context) {
			model := c.PostForm("model")
			if model == "" {
				c.String(http.StatusBadRequest, "Table name is required")
				return
			}

			db.Connect()

			sqlQuery := fmt.Sprintf("SELECT * FROM \"%s\"", model)

			rows, err := db.Query(sqlQuery)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to get data: %s", err.Error())
				return
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to get columns: %s", err.Error())
				return
			}

			var results [][]interface{}
			for rows.Next() {
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range columns {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					c.String(http.StatusInternalServerError, "Failed to scan row: %s", err.Error())
					return
				}

				// Convert byte slices to strings for display
				for i, val := range values {
					if b, ok := val.([]byte); ok {
						values[i] = string(b)
					}
				}
				results = append(results, values)
			}

			c.HTML(http.StatusOK, "admin/data.html", gin.H{
				"Title":   "Table Data",
				"Model":   model,
				"Columns": columns,
				"Rows":    results,
			})
		})
	}

	return router
}

func main() {
	db = s.Database{}
	db.Config = s.DatabaseConfig{}
	db.Config.Host = "127.0.0.1"
	db.Config.Port = 5432
	db.Config.User = "postgres"
	db.Config.Password = "postgres"
	db.Config.Database = "postgres"
	db.Config.SSL = "disable"
	db.Config.Driver = s.Driver(s.POSTGRESQL)

	db.Connect()
	defer db.Close()

	router := Router()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadTimeout:       60 * time.Second,
		IdleTimeout:       15 * time.Minute,
	}

	if err := srv.ListenAndServe(); err != nil {
		u.ErrorF("Failed to start server!:\t%s\n", err.Error())
	}
}
