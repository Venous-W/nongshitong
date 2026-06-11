package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"nongshaitong/internal/db"
	"nongshaitong/internal/handler"

	"github.com/gin-gonic/gin"
)

// web 目录下的所有文件通过 embed 打包进二进制，发布时只需一个 .exe 文件。
//
//go:embed web
var webFS embed.FS

func main() {
	// 初始化数据库（首次运行自动创建 data.db 并写入种子数据）
	db.Init("data.db")

	// 生产模式下关闭 Gin 的调试日志
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// ── 注册 API 路由 ──────────────────────────────────────────────────────────
	api := r.Group("/api")
	{
		// 分类相关
		api.GET("/categories", handler.GetCategoryTree)    // 树形（前端查询页用）
		api.GET("/categories/flat", handler.GetCategories) // 平铺（管理台下拉用）
		api.POST("/categories", handler.CreateCategory)
		api.PUT("/categories/:id", handler.UpdateCategory)
		api.DELETE("/categories/:id", handler.DeleteCategory)

		// 作物/场所相关
		api.GET("/crops", handler.GetAllCrops)
		api.POST("/crops", handler.CreateCrop)
		api.PUT("/crops/:id", handler.UpdateCrop)
		api.DELETE("/crops/:id", handler.DeleteCrop)

		// 防治对象相关
		api.GET("/targets", handler.GetAllTargets)
		api.POST("/targets", handler.CreateTarget)
		api.PUT("/targets/:id", handler.UpdateTarget)
		api.DELETE("/targets/:id", handler.DeleteTarget)

		// 动态筛选（查询页按钮联动）
		api.GET("/filter/crops", handler.GetFilterCrops)
		api.GET("/filter/targets", handler.GetFilterTargets)

		// 产品相关
		api.GET("/products", handler.ListProducts)
		api.GET("/products/:id", handler.GetProductDetail)
		api.POST("/products", handler.CreateProduct)
		api.PUT("/products/:id", handler.UpdateProduct)
		api.DELETE("/products/:id", handler.DeleteProduct)
		api.PUT("/products/:id/active", handler.ToggleActive)
	}

	// ── 注册前端页面路由 ────────────────────────────────────────────────────────
	// 从 embed FS 中提取 web 子目录
	webSubFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("加载前端文件失败: %v", err)
	}

	// 创建 HTTP 文件服务器
	webHttpFS := http.FS(webSubFS)
	fileServer := http.FileServer(webHttpFS)

	// 管理页：路径含 /manage 时返回 admin.html
	r.GET("/manage", func(c *gin.Context) {
		// 读取 admin.html 内容并直接返回
		content, err := fs.ReadFile(webSubFS, "admin.html")
		if err != nil {
			c.String(404, "File not found: admin.html")
			return
		}
		c.Data(200, "text/html; charset=utf-8", content)
	})
	r.GET("/manage/*path", func(c *gin.Context) {
		// 读取 admin.html 内容并直接返回
		content, err := fs.ReadFile(webSubFS, "admin.html")
		if err != nil {
			c.String(404, "File not found: admin.html")
			return
		}
		c.Data(200, "text/html; charset=utf-8", content)
	})

	// 根路径：返回 index.html
	r.GET("/", func(c *gin.Context) {
		// 读取 index.html 内容并直接返回
		content, err := fs.ReadFile(webSubFS, "index.html")
		if err != nil {
			c.String(404, "File not found: index.html")
			return
		}
		c.Data(200, "text/html; charset=utf-8", content)
	})

	// 处理所有其他路径（静态资源和前端路由）
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是API请求，返回404
		if strings.HasPrefix(path, "/api/") {
			c.JSON(404, gin.H{"code": 404, "msg": "API endpoint not found"})
			return
		}

		// 如果是管理页面相关的路径，返回admin.html
		if strings.HasPrefix(path, "/manage") {
			content, err := fs.ReadFile(webSubFS, "admin.html")
			if err != nil {
				c.String(404, "File not found: admin.html")
				return
			}
			c.Data(200, "text/html; charset=utf-8", content)
			return
		}

		// 检查请求的文件是否存在（移除开头的斜杠）
		requestedFile := strings.TrimPrefix(path, "/")
		if requestedFile == "" {
			requestedFile = "index.html"
		}

		// 检查文件是否存在
		if _, err := fs.Stat(webSubFS, requestedFile); err == nil {
			// 文件存在，使用文件服务器返回
			c.Request.URL.Path = "/" + requestedFile
			fileServer.ServeHTTP(c.Writer, c.Request)
		} else {
			// 文件不存在，返回 index.html（用于前端路由）
			content, err := fs.ReadFile(webSubFS, "index.html")
			if err != nil {
				c.String(404, "File not found: index.html")
				return
			}
			c.Data(200, "text/html; charset=utf-8", content)
		}
	})

	log.Println("服务启动：http://localhost:9000")
	log.Println("查询页：  http://localhost:9000/")
	log.Println("管理页：  http://localhost:9000/manage")

	if err = r.Run(":9000"); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
