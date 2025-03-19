package mysql

import (
	"context"
	"fmt"
	"honeypot/server/pkg/logger"
	"time"

	"github.com/src-d/go-mysql-server"
	"github.com/src-d/go-mysql-server/auth"
	"github.com/src-d/go-mysql-server/memory"
	"github.com/src-d/go-mysql-server/server"
	"github.com/src-d/go-mysql-server/sql"
)

// StartMySQL 启动MySQL服务
func StartMySQL(addr string, isProxy bool) error {
	// 修改后端 MySQL 端口
	backendAddr := "127.0.0.1:3367"  // 改为 3367
	frontendAddr := addr

	// 启动后端MySQL服务
	go func() {
		if err := startBackendServer(backendAddr); err != nil {
			logger.Log.Errorf("Backend MySQL server failed: %v", err)
		}
	}()

	// 启动代理服务
	proxy := NewMySQLProxy(frontendAddr, backendAddr, isProxy)
	return proxy.Start()
}

// startBackendServer 启动后端MySQL服务
func startBackendServer(addr string) error {
	engine := sqle.NewDefault()
	engine.AddDatabase(createTestDatabase())

	// 添加所有诱饵数据库
	for _, db := range createDecoyDatabases() {
		engine.AddDatabase(db)
	}

	config := server.Config{
		Protocol: "tcp",
		Address:  addr,
		Auth:     auth.NewNativeSingle("root", "123456", auth.AllPermissions),
	}

	s, err := server.NewDefaultServer(config, engine)
	if err != nil {
		return fmt.Errorf("cannot create server: %v", err)
	}

	return s.Start()
}

// createTestDatabase 创建测试数据库
func createTestDatabase() *memory.Database {
	const (
		dbName    = "my_db"
		tableName = "my_table"
	)

	db := memory.NewDatabase(dbName)
	table := memory.NewTable(tableName, sql.Schema{
		{Name: "name", Type: sql.Text, Nullable: false, Source: tableName},
		{Name: "email", Type: sql.Text, Nullable: false, Source: tableName},
		{Name: "phone_numbers", Type: sql.Text, Nullable: false, Source: tableName},
		{Name: "created_at", Type: sql.Timestamp, Nullable: false, Source: tableName},
	})

	db.AddTable(tableName, table)
	ctx := sql.NewContext(context.Background())

	// 添加示例数据并处理错误
	if err := table.Insert(ctx, sql.NewRow("admin", "admin@example.com", "123456789", time.Now())); err != nil {
		logger.Log.Errorf("Failed to insert admin record: %v", err)
	}
	if err := table.Insert(ctx, sql.NewRow("user1", "user1@example.com", "987654321", time.Now())); err != nil {
		logger.Log.Errorf("Failed to insert user1 record: %v", err)
	}

	return db
}

// createDecoyDatabases 创建诱饵数据库
func createDecoyDatabases() []*memory.Database {
	decoyDbs := []struct {
		name   string
		tables []string
	}{
		{"admin_db", []string{"users", "permissions"}},
		{"web_app", []string{"accounts", "orders", "products"}},
		{"backup_db", []string{"backup_logs", "system_config"}},
	}

	var databases []*memory.Database

	for _, db := range decoyDbs {
		database := memory.NewDatabase(db.name)
		for _, tableName := range db.tables {
			table := memory.NewTable(tableName, sql.Schema{
				{Name: "id", Type: sql.Int32, Nullable: false, Source: tableName},
				{Name: "data", Type: sql.Text, Nullable: false, Source: tableName},
			})
			database.AddTable(tableName, table)
		}
		databases = append(databases, database)
	}

	return databases
}
