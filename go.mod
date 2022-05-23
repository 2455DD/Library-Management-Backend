module lms

go 1.15

require (
	github.com/boombuler/barcode v1.0.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/contrib v0.0.0-20201101042839-6a891bf89f19
	github.com/gin-gonic/gin v1.7.7
	github.com/go-ini/ini v1.66.4
	github.com/smartwalle/alipay/v3 v3.1.7
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/image v0.0.0-20220413100746-70e8d0d3baa9
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gorm.io/driver/mysql v1.3.3
	gorm.io/gorm v1.23.4
)

replace github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt v3.2.2+incompatible
