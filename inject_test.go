package inject

import (
	"encoding/json"
	"fmt"
	"testing"
)

type User struct {
	Name     string
	Age      int
	Address  *Address  `inject`
	AuthCoed *AuthCode `inject`
}

type Address struct {
	Province string
	City     string
	Area     string
	Street   string
}

type AuthCode struct {
	Code string
}

func NewAddress(province, city, area, street string) *Address {
	//return func() (interface{}, error) {
	return &Address{
		Province: province,
		City:     city,
		Area:     area,
		Street:   street,
	}
	//}
}

func TestInject(t *testing.T) {
	var user User
	injector := New()
	auth := NewAddress("广东", "深圳", "宝安", "新安")
	injector.Map(&AuthCode{Code: "123456"})
	injector.Map(auth)
	//inject.Invoke(func(authCode *AuthCode, address *Address) {
	//	user.AuthCoed = authCode
	//	user.Address = address
	//})
	fmt.Println(injector.Apply(&user))
	bytes, _ := json.Marshal(user)
	fmt.Println(string(bytes))
}
