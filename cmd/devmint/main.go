package devmint

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/utils"
	"gorm.io/gorm"
)

var db *gorm.DB
var jwtSvc *utils.JWTService

func mintHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("APP_ENV") != "development" {
		http.Error(w, "disabled", 403)
		return
	}
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "superadmin"
	}
	var u model.User
	if err := db.Preload("Role").Where("username = ?", username).First(&u).Error; err != nil {
		http.Error(w, "user not found: "+err.Error(), 404)
		return
	}
	access, err := jwtSvc.GenerateAccessToken(u.ID, u.RoleID, u.Role.Name)
	if err != nil {
		http.Error(w, "issue: "+err.Error(), 500)
		return
	}
	refresh, _, _ := jwtSvc.GenerateRefreshToken(u.ID)
	resp := map[string]any{
		"success": true,
		"data": map[string]any{
			"access_token":  access,
			"refresh_token": refresh,
			"user": map[string]any{
				"id":             u.ID,
				"username":       u.Username,
				"email":          u.Email,
				"full_name":      u.FullName,
				"role_name":      u.Role.Name,
				"is_2fa_enabled": u.Is2FAEnabled,
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	db = config.NewDatabase(cfg)
	jwtSvc = utils.NewJWTService(&cfg.JWT)

	http.HandleFunc("/dev/mint-token", mintHandler)
	addr := ":9099"
	log.Printf("Dev token minter on %s", addr)
	fmt.Println(http.ListenAndServe(addr, nil))
}
