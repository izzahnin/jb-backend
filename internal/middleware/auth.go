package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims adalah payload yang di-encode dalam JWT token.
// Berisi informasi user yang akan di-verify di setiap request.
type Claims struct {
	UserID   int    `json:"user_id"`   // ID user dari database
	Username string `json:"username"` // Username untuk logging/audit
	Role     string `json:"role"`     // 'admin' atau 'customer' untuk authorization
	jwt.RegisteredClaims
}

// AuthMiddleware adalah middleware untuk memvalidasi JWT token dari request header.
// Digunakan pada protected routes (mis. /admin/*).
// Jika token valid, claim data di-inject ke gin.Context untuk dipakai handler berikutnya.
// Jika token invalid/missing, return 401 Unauthorized.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. EXTRACT: Ambil Authorization header dari request
		// Format: "Authorization: Bearer <token>"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort() // Stop request chain, jangan lanjut ke handler
			return
		}

		// 2. PARSE: Split header untuk dapatkan token (remove "Bearer " prefix)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 3. VERIFY: Parse JWT token dan validasi signature
		// jwtSecret digunakan untuk verify bahwa token asli dari server (bukan dipalsukan)
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Verify algorithm adalah HS256 (untuk keamanan, reject algoritma lain)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Return secret key untuk verify signature
			return []byte(jwtSecret), nil
		})

		// 4. CHECK: Verifikasi token valid (signature ok + tidak expired)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// 5. INJECT: Simpan claims ke gin.Context agar bisa diakses handler selanjutnya
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		// 6. PROCEED: Lanjutkan ke handler berikutnya
		c.Next()
	}
}

// AdminMiddleware adalah middleware untuk check apakah user punya role 'admin'.
// Harus dipanggil SETELAH AuthMiddleware (agar sudah ada user claims di context).
// Digunakan pada /admin/* routes untuk mencegah customer access.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role dari context (set by AuthMiddleware)
		role, exists := c.Get("role")
		if !exists {
			// Jika role tidak ada, berarti AuthMiddleware gagal (tidak seharusnya terjadi)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing role in context"})
			c.Abort()
			return
		}

		// Check apakah role adalah 'admin'
		roleStr, ok := role.(string)
		if !ok || roleStr != "admin" {
			// Jika bukan admin, return 403 Forbidden (authenticated tapi tidak authorized)
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			c.Abort()
			return
		}

		// Proceed ke handler
		c.Next()
	}
}
