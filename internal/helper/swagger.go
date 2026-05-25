package helper

import (
	"github.com/gofiber/fiber/v3"
)

// ServeSwaggerUI mengembalikan halaman HTML Swagger UI yang terhubung ke spesifikasi API kita.
func ServeSwaggerUI(c fiber.Ctx) error {
	c.Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Go Fiber v3 Auth - Dokumentasi API Swagger</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = () => {
            window.ui = SwaggerUIBundle({
                url: '/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "BaseLayout"
            });
        };
    </script>
</body>
</html>`
	return c.SendString(html)
}

// ServeSwaggerJSON mengembalikan data JSON mentah untuk dokumen spesifikasi OpenAPI 3.0.
func ServeSwaggerJSON(c fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	return c.SendString(SwaggerJSON)
}

// SwaggerJSON adalah spesifikasi OpenAPI 3.0 lengkap untuk sistem Auth, RBAC, dan Menu ini.
const SwaggerJSON = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Go Fiber v3 Auth & RBAC API",
    "description": "Dokumentasi API lengkap untuk layanan Autentikasi, Refresh Token, Role-Based Access Control (RBAC) dan Manajemen Menu Dinamis.",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Server Pengembangan Lokal"
    }
  ],
  "paths": {
    "/auth/register": {
      "post": {
        "tags": ["Autentikasi"],
        "summary": "Pendaftaran Akun Baru",
        "description": "Mendaftarkan pengguna baru ke sistem dengan role bawaan 'user'.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name", "email", "password"],
                "properties": {
                  "name": { "type": "string", "example": "Budi Santoso" },
                  "email": { "type": "string", "format": "email", "example": "budi@example.com" },
                  "password": { "type": "string", "minLength": 8, "example": "password123" }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Registrasi Berhasil",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "message": { "type": "string", "example": "registered successfully" }
                  }
                }
              }
            }
          },
          "400": { "description": "Payload tidak valid" },
          "409": { "description": "Email sudah terdaftar" }
        }
      }
    },
    "/auth/login": {
      "post": {
        "tags": ["Autentikasi"],
        "summary": "Masuk / Login Pengguna",
        "description": "Memvalidasi kredensial pengguna dan mengembalikan Access Token (JWT) serta Refresh Token.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["email", "password"],
                "properties": {
                  "email": { "type": "string", "format": "email", "example": "budi@example.com" },
                  "password": { "type": "string", "example": "password123" }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Login Berhasil",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/TokenResponse"
                }
              }
            }
          },
          "401": { "description": "Kredensial tidak valid" },
          "403": { "description": "Akun tidak aktif" }
        }
      }
    },
    "/auth/refresh": {
      "post": {
        "tags": ["Autentikasi"],
        "summary": "Refresh Access Token",
        "description": "Menukarkan Refresh Token yang valid untuk mendapatkan Access Token dan Refresh Token baru (Rotasi Sesi).",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["refresh_token"],
                "properties": {
                  "refresh_token": { "type": "string", "example": "a671cfb9..." }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Token Berhasil Diperbarui",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/TokenResponse"
                }
              }
            }
          },
          "401": { "description": "Refresh Token tidak valid atau kadaluwarsa" }
        }
      }
    },
    "/auth/logout": {
      "post": {
        "tags": ["Autentikasi"],
        "summary": "Keluar / Logout Single Sesi",
        "description": "Membatalkan/merevoke satu Refresh Token tertentu.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["refresh_token"],
                "properties": {
                  "refresh_token": { "type": "string", "example": "a671cfb9..." }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Logout Berhasil",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "message": { "type": "string", "example": "logged out" }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/profile": {
      "get": {
        "tags": ["Pengguna Pribadi"],
        "summary": "Mendapatkan Profil Pengguna Saat Ini",
        "description": "Mengembalikan data identitas pengguna yang terenkripsi di dalam JWT.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "Data Profil Ditemukan",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "userId": { "type": "string", "example": "d3b07384-d113-4956-a5db-e17fcd506030" },
                    "name": { "type": "string", "example": "Budi Santoso" },
                    "role": { "type": "string", "example": "user" },
                    "roles": { "type": "array", "items": { "type": "string" }, "example": ["user"] }
                  }
                }
              }
            }
          },
          "401": { "description": "Tidak diizinkan / Token tidak valid" }
        }
      }
    },
    "/api/logout-all": {
      "post": {
        "tags": ["Autentikasi"],
        "summary": "Keluar dari Semua Perangkat (Logout All)",
        "description": "Membatalkan seluruh sesi aktif (semua Refresh Token) milik pengguna saat ini.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "Semua sesi berhasil ditutup",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "message": { "type": "string", "example": "all sessions terminated" }
                  }
                }
              }
            }
          },
          "401": { "description": "Tidak diizinkan" }
        }
      }
    },
    "/api/menu": {
      "get": {
        "tags": ["Menu Navigasi"],
        "summary": "Mendapatkan Struktur Menu Navigasi Pengguna",
        "description": "Mengembalikan pohon hierarki menu (sidebar) yang diizinkan berdasarkan hak akses role pengguna saat ini.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "Struktur Menu Berhasil Dibuat",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "menus": {
                      "type": "array",
                      "items": { "$ref": "#/components/schemas/MenuResponse" }
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/admin/users": {
      "get": {
        "tags": ["Manajemen Pengguna (Admin)"],
        "summary": "Daftar Pengguna Terpaginasi",
        "description": "Mendapatkan daftar seluruh pengguna. Admin hanya dapat melihat pengguna dengan role 'user', sedangkan Super Admin dapat melihat semuanya.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "role", "in": "query", "schema": { "type": "string" }, "description": "Filter berdasarkan nama role (Super Admin saja)" },
          { "name": "is_active", "in": "query", "schema": { "type": "boolean" }, "description": "Filter status aktif" },
          { "name": "search", "in": "query", "schema": { "type": "string" }, "description": "Pencarian nama atau email" },
          { "name": "page", "in": "query", "schema": { "type": "integer", "default": 1 } },
          { "name": "page_size", "in": "query", "schema": { "type": "integer", "default": 20 } }
        ],
        "responses": {
          "200": {
            "description": "Daftar Pengguna Berhasil Diambil",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "data": { "type": "array", "items": { "$ref": "#/components/schemas/UserResponse" } },
                    "total": { "type": "integer", "example": 1 },
                    "page": { "type": "integer", "example": 1 },
                    "page_size": { "type": "integer", "example": 20 },
                    "total_pages": { "type": "integer", "example": 1 }
                  }
                }
              }
            }
          }
        }
      },
      "post": {
        "tags": ["Manajemen Pengguna (Admin)"],
        "summary": "Membuat Pengguna Baru",
        "description": "Admin hanya diperbolehkan membuat pengguna dengan role 'user'. Super Admin bebas menentukan role.",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name", "email", "password", "role"],
                "properties": {
                  "name": { "type": "string", "example": "Admin Junior" },
                  "email": { "type": "string", "format": "email", "example": "jr.admin@example.com" },
                  "password": { "type": "string", "example": "password123" },
                  "role": { "type": "string", "enum": ["super_admin", "admin", "user"], "example": "user" }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Pengguna Berhasil Dibuat",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/UserResponse" }
              }
            }
          },
          "403": { "description": "Melanggar batasan RBAC" }
        }
      }
    },
    "/api/admin/users/{id}": {
      "get": {
        "tags": ["Manajemen Pengguna (Admin)"],
        "summary": "Mendapatkan Detail Pengguna",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }],
        "responses": {
          "200": {
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/UserResponse" } } }
          }
        }
      },
      "patch": {
        "tags": ["Manajemen Pengguna (Admin)"],
        "summary": "Memperbarui Profil Pengguna",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }],
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "name": { "type": "string", "example": "Nama Baru" },
                  "email": { "type": "string", "format": "email" },
                  "password": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/UserResponse" } } }
          }
        }
      },
      "delete": {
        "tags": ["Manajemen Pengguna (Admin)"],
        "summary": "Menghapus Pengguna (Soft Delete)",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }],
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "message": { "type": "string", "example": "user deleted" }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/admin/users/{id}/active": {
      "patch": {
        "tags": ["Manajemen Pengguna (Admin)"],
        "summary": "Mengubah Status Aktif Pengguna",
        "description": "Mengaktifkan atau menonaktifkan akun pengguna.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["is_active"],
                "properties": {
                  "is_active": { "type": "boolean", "example": false }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "message": { "type": "string", "example": "user deactivated" }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/super-admin/users/{id}/role": {
      "patch": {
        "tags": ["Manajemen Pengguna (Super Admin)"],
        "summary": "Mengubah Role Pengguna",
        "description": "Hanya diizinkan diakses oleh Super Admin untuk mengatur ulang hak akses pengguna.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["role"],
                "properties": {
                  "role": { "type": "string", "enum": ["super_admin", "admin", "user"], "example": "admin" }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/UserResponse" } } }
          }
        }
      }
    },
    "/api/super-admin/roles": {
      "get": {
        "tags": ["Manajemen Role & Menu (Super Admin)"],
        "summary": "Katalog Role Sistem",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "properties": {
                      "name": { "type": "string" },
                      "description": { "type": "string" }
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/super-admin/menus": {
      "get": {
        "tags": ["Manajemen Role & Menu (Super Admin)"],
        "summary": "Daftar Flat Semua Menu",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "menus": { "type": "array", "items": { "$ref": "#/components/schemas/MenuResponse" } }
                  }
                }
              }
            }
          }
        }
      },
      "post": {
        "tags": ["Manajemen Role & Menu (Super Admin)"],
        "summary": "Membuat Item Menu Baru",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name", "key", "path"],
                "properties": {
                  "name": { "type": "string", "example": "Manajemen Produk" },
                  "key": { "type": "string", "example": "product-management" },
                  "path": { "type": "string", "example": "/admin/products" },
                  "icon": { "type": "string", "example": "package" },
                  "parent_id": { "type": "string", "format": "uuid", "nullable": true },
                  "sort_order": { "type": "integer", "example": 5 },
                  "is_active": { "type": "boolean", "default": true },
                  "role_keys": { "type": "array", "items": { "type": "string" }, "example": ["super_admin", "admin"] }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MenuResponse" } } }
          }
        }
      }
    },
    "/api/super-admin/menus/tree": {
      "get": {
        "tags": ["Manajemen Role & Menu (Super Admin)"],
        "summary": "Mendapatkan Pohon Hierarki Menu Lengkap",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "menus": { "type": "array", "items": { "$ref": "#/components/schemas/MenuResponse" } }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/super-admin/menus/{id}": {
      "get": {
        "tags": ["Manajemen Role & Menu (Super Admin)"],
        "summary": "Mendapatkan Detail Menu",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }],
        "responses": {
          "200": {
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MenuResponse" } } }
          }
        }
      },
      "patch": {
        "tags": ["Manajemen Role & Menu (Super Admin)"],
        "summary": "Memperbarui Data Item Menu",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }],
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "name": { "type": "string" },
                  "path": { "type": "string" },
                  "icon": { "type": "string" },
                  "sort_order": { "type": "integer" },
                  "is_active": { "type": "boolean" }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MenuResponse" } } }
          }
        }
      },
      "delete": {
        "tags": ["Manajemen Role & Menu (Super Admin)"],
        "summary": "Menghapus Item Menu",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }],
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "message": { "type": "string", "example": "menu deleted" }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/super-admin/menus/{id}/roles": {
      "put": {
        "tags": ["Manajemen Role & Menu (Super Admin)"],
        "summary": "Menyusun Ulang Hak Akses Role pada Menu",
        "description": "Mengganti hak akses role yang diizinkan membuka menu ini.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["role_keys"],
                "properties": {
                  "role_keys": { "type": "array", "items": { "type": "string" }, "example": ["super_admin", "admin"] }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MenuResponse" } } }
          }
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "Masukkan Token JWT Anda dengan format: 'Bearer <token>'"
      }
    },
    "schemas": {
      "TokenResponse": {
        "type": "object",
        "properties": {
          "access_token": { "type": "string", "example": "eyJhbGciOi..." },
          "refresh_token": { "type": "string", "example": "a671cfb9..." },
          "token_type": { "type": "string", "example": "Bearer" },
          "expires_in": { "type": "integer", "example": 900 }
        }
      },
      "UserResponse": {
        "type": "object",
        "properties": {
          "id": { "type": "string", "format": "uuid", "example": "d3b07384-d113-4956-a5db-e17fcd506030" },
          "name": { "type": "string", "example": "Budi Santoso" },
          "email": { "type": "string", "format": "email", "example": "budi@example.com" },
          "is_active": { "type": "boolean", "example": true },
          "roles": { "type": "array", "items": { "type": "string" }, "example": ["user"] }
        }
      },
      "MenuResponse": {
        "type": "object",
        "properties": {
          "id": { "type": "string", "format": "uuid", "example": "e58129ac-c4b9-4d6c-bbcb-df12eaef9321" },
          "name": { "type": "string", "example": "User Management" },
          "key": { "type": "string", "example": "user-management" },
          "path": { "type": "string", "example": "/admin/users" },
          "icon": { "type": "string", "example": "users" },
          "parent_id": { "type": "string", "format": "uuid", "nullable": true, "example": null },
          "sort_order": { "type": "integer", "example": 2 },
          "is_active": { "type": "boolean", "example": true },
          "roles": { "type": "array", "items": { "type": "string" }, "example": ["super_admin", "admin"] },
          "children": { "type": "array", "items": { "$ref": "#/components/schemas/MenuResponse" } }
        }
      }
    }
  }
}`
