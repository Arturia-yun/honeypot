package mysql

import (
	"context"
	"fmt"
	"honeypot/server/pkg/logger"
	"net"
	"time"
	"os"
	"strings"
	"github.com/src-d/go-mysql-server"
	"github.com/src-d/go-mysql-server/auth"
	"github.com/src-d/go-mysql-server/memory"
	"github.com/src-d/go-mysql-server/server"
	"github.com/src-d/go-mysql-server/sql"
)

// StartMySQL 启动MySQL服务
// 在现有代码中添加文件读取处理功能

// 在现有代码的基础上添加以下函数

// LoadDictionary 加载文件读取字典
func LoadDictionary(dictPath string) ([]string, error) {
	// 如果未指定字典路径，使用默认路径
	if dictPath == "" {
		dictPath = "pkg/service/MySQL/dicc.txt"
	}
	
	// 读取字典文件
	content, err := os.ReadFile(dictPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取字典文件: %v", err)
	}
	
	// 按行分割
	lines := strings.Split(string(content), "\n")
	
	// 过滤空行
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	
	// 使用结构化日志
	logger.Log.Infof("已加载 %d 个文件路径到字典", len(result))
	
	// 记录字典加载事件
	dictLog := struct {
		Time      time.Time `json:"time"`
		Service   string    `json:"service"`
		Event     string    `json:"event"`
		DictPath  string    `json:"dict_path"`
		EntryCount int      `json:"entry_count"`
	}{
		Time:       time.Now(),
		Service:    "mysql",
		Event:      "dict_loaded",
		DictPath:   dictPath,
		EntryCount: len(result),
	}
	
	logger.LogReport.WithField("api", "/api/mysql/").Info(dictLog)
	
	return result, nil
}

// 修改 StartMySQL 函数，添加字典加载
func StartMySQL(addr string, isProxy bool) error {
	// 修改后端 MySQL 端口
	backendAddr := "127.0.0.1:3367"  // 改为 3367
	frontendAddr := addr
	
	// 加载文件读取字典
	_, err := LoadDictionary("")
	if err != nil {
		logger.Log.Warningf("加载文件读取字典失败: %v", err)
	}
	
	// 启动后端MySQL服务
	go func() {
		if err := startBackendServer(backendAddr); err != nil {
			logger.Log.Errorf("Backend MySQL server failed: %v", err)
		}
	}()
	
	// 启动文件读取漏洞监听器
	go func() {
		fileReadAddr := "0.0.0.0:3307"  // 在3307端口监听文件读取请求
		listener, err := net.Listen("tcp", fileReadAddr)
		if err != nil {
			logger.Log.Errorf("MySQL file reader listener failed: %v", err)
			return
		}
		
		logger.Log.Warningf("MySQL file reader listening on %s", fileReadAddr)
		
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Log.Errorf("Accept connection error: %v", err)
				continue
			}
			
			go FileReadHandler(conn, isProxy)
		}
	}()
	
	// 启动MySQL代理
	proxy := &MySQLProxy{
		frontendAddr: frontendAddr,
		backendAddr:  backendAddr,
		isProxy:      isProxy,
	}
	
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
