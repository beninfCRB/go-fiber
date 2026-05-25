package database

import (
	"backend/internal/models"
	"log"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.RoleModel{},
		&models.User{},
		&models.RefreshToken{},
		&models.Menu{},
		&models.AuditLog{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("database migration completed")

	seedRoles(db)
	seedMenus(db)
}

// seedRoles inserts the three base roles if they don't exist.
func seedRoles(db *gorm.DB) {
	roles := []models.RoleModel{
		{Name: models.RoleSuperAdmin, Description: "Full system access"},
		{Name: models.RoleAdmin, Description: "Tenant/organization admin"},
		{Name: models.RoleUser, Description: "Regular user"},
	}
	for _, r := range roles {
		db.FirstOrCreate(&models.RoleModel{}, models.RoleModel{Name: r.Name, Description: r.Description})
	}
	log.Println("roles seeded")
}

// seedMenus inserts the default menu structure and assigns roles.
// This runs idempotently (uses FirstOrCreate on the unique Key field).
func seedMenus(db *gorm.DB) {
	// Fetch roles for assignment
	var superAdmin, admin, user models.RoleModel
	db.Where("name = ?", models.RoleSuperAdmin).First(&superAdmin)
	db.Where("name = ?", models.RoleAdmin).First(&admin)
	db.Where("name = ?", models.RoleUser).First(&user)

	allRoles := []models.RoleModel{superAdmin, admin, user}
	adminRoles := []models.RoleModel{superAdmin, admin}
	superOnly := []models.RoleModel{superAdmin}

	type menuSeed struct {
		menu  models.Menu
		roles []models.RoleModel
	}

	// ── Root menus ────────────────────────────────────────────────────────────
	roots := []menuSeed{
		{
			menu: models.Menu{
				Name: "Dashboard", Key: "dashboard",
				Path: "/dashboard", Icon: "home", SortOrder: 1, IsActive: true,
			},
			roles: allRoles,
		},
		{
			menu: models.Menu{
				Name: "User Management", Key: "user-management",
				Path: "/admin/users", Icon: "users", SortOrder: 2, IsActive: true,
			},
			roles: adminRoles,
		},
		{
			menu: models.Menu{
				Name: "Menu Management", Key: "menu-management",
				Path: "/super-admin/menus", Icon: "menu", SortOrder: 3, IsActive: true,
			},
			roles: superOnly,
		},
		{
			menu: models.Menu{
				Name: "Role Management", Key: "role-management",
				Path: "/super-admin/roles", Icon: "shield", SortOrder: 4, IsActive: true,
			},
			roles: superOnly,
		},
		{
			menu: models.Menu{
				Name: "Profile", Key: "profile",
				Path: "/profile", Icon: "user", SortOrder: 10, IsActive: true,
			},
			roles: allRoles,
		},
	}

	for _, seed := range roots {
		var existing models.Menu
		result := db.Where("key = ?", seed.menu.Key).First(&existing)
		if result.Error != nil {
			// Create new
			db.Create(&seed.menu)
			db.Model(&seed.menu).Association("Roles").Replace(seed.roles)
		}
		// If already exists, skip to keep manual changes intact
	}

	// ── Sub-menus under "User Management" ────────────────────────────────────
	var userMgmtMenu models.Menu
	if err := db.Where("key = ?", "user-management").First(&userMgmtMenu).Error; err == nil {
		subMenus := []menuSeed{
			{
				menu: models.Menu{
					Name: "All Users", Key: "users-list",
					Path: "/admin/users", Icon: "list", ParentID: &userMgmtMenu.ID,
					SortOrder: 1, IsActive: true,
				},
				roles: adminRoles,
			},
			{
				menu: models.Menu{
					Name: "Manage Admins", Key: "users-admins",
					Path: "/super-admin/users?role=admin", Icon: "shield-check",
					ParentID: &userMgmtMenu.ID, SortOrder: 2, IsActive: true,
				},
				roles: superOnly,
			},
		}
		for _, seed := range subMenus {
			var existing models.Menu
			result := db.Where("key = ?", seed.menu.Key).First(&existing)
			if result.Error != nil {
				db.Create(&seed.menu)
				db.Model(&seed.menu).Association("Roles").Replace(seed.roles)
			}
		}
	}

	log.Println("menus seeded")
}
