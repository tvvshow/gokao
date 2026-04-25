package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"unicode"
)

// PasswordValidator 密码验证器
type PasswordValidator struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
	MaxLength      int
	CommonPasswords map[string]bool
}

// NewPasswordValidator 创建密码验证器
func NewPasswordValidator(minLength int, requireUpper, requireLower, requireNumber, requireSpecial bool) *PasswordValidator {
	return &PasswordValidator{
		MinLength:      minLength,
		RequireUpper:   requireUpper,
		RequireLower:   requireLower,
		RequireNumber:  requireNumber,
		RequireSpecial: requireSpecial,
		MaxLength:      128,
		CommonPasswords: loadCommonPasswords(),
	}
}

// ValidatePassword 验证密码复杂度
func (v *PasswordValidator) ValidatePassword(password string) error {
	if len(password) < v.MinLength {
		return fmt.Errorf("password must be at least %d characters long", v.MinLength)
	}

	if len(password) > v.MaxLength {
		return fmt.Errorf("password must not exceed %d characters", v.MaxLength)
	}

	if v.CommonPasswords[strings.ToLower(password)] {
		return errors.New("password is too common and easily guessable")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	var validationErrors []string

	if v.RequireUpper && !hasUpper {
		validationErrors = append(validationErrors, "at least one uppercase letter")
	}

	if v.RequireLower && !hasLower {
		validationErrors = append(validationErrors, "at least one lowercase letter")
	}

	if v.RequireNumber && !hasNumber {
		validationErrors = append(validationErrors, "at least one number")
	}

	if v.RequireSpecial && !hasSpecial {
		validationErrors = append(validationErrors, "at least one special character")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("password must contain %s", strings.Join(validationErrors, ", "))
	}

	if v.hasRepeatingPattern(password) {
		return errors.New("password contains easily guessable patterns")
	}

	return nil
}

// ValidatePasswordStrength 验证密码强度并返回强度分数 (0-100)
func (v *PasswordValidator) ValidatePasswordStrength(password string) (int, error) {
	if err := v.ValidatePassword(password); err != nil {
		return 0, err
	}

	var score int

	// 长度分数
	length := len(password)
	if length >= 12 {
		score += 25
	} else if length >= 8 {
		score += 15
	} else {
		score += 5
	}

	// 字符类型多样性分数
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	charTypeCount := 0
	if hasUpper {
		charTypeCount++
		score += 10
	}
	if hasLower {
		charTypeCount++
		score += 10
	}
	if hasNumber {
		charTypeCount++
		score += 10
	}
	if hasSpecial {
		charTypeCount++
		score += 15
	}

	// 额外多样性奖励
	if charTypeCount >= 3 {
		score += 10
	}

	// 检查常见模式并扣分
	if v.hasRepeatingPattern(password) {
		score -= 20
	}

	if v.CommonPasswords[strings.ToLower(password)] {
		score = 0
	}

	// 确保分数在0-100范围内
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score, nil
}

// GetPasswordStrengthLabel 获取密码强度标签
func (v *PasswordValidator) GetPasswordStrengthLabel(score int) string {
	switch {
	case score >= 80:
		return "very_strong"
	case score >= 60:
		return "strong"
	case score >= 40:
		return "medium"
	case score >= 20:
		return "weak"
	default:
		return "very_weak"
	}
}

// hasRepeatingPattern 检查密码中是否有重复模式
func (v *PasswordValidator) hasRepeatingPattern(password string) bool {
	// 检查连续字符 (如 "aaaa", "1234", "abcd")
	if v.hasConsecutiveChars(password, 4) {
		return true
	}

	// 检查键盘布局模式 (如 "qwerty", "asdfgh")
	keyboardPatterns := []string{
		"qwerty", "asdfgh", "zxcvbn", "123456", "abcdef",
		"!@#$%^", "password", "admin123", "welcome", "letmein",
	}

	lowerPassword := strings.ToLower(password)
	for _, pattern := range keyboardPatterns {
		if strings.Contains(lowerPassword, pattern) {
			return true
		}
	}

	// 检查重复子串
	if v.hasRepeatingSubstring(password, 3) {
		return true
	}

	return false
}

// hasConsecutiveChars 检查是否有连续字符
func (v *PasswordValidator) hasConsecutiveChars(s string, maxConsecutive int) bool {
	if len(s) < maxConsecutive {
		return false
	}

	for i := 0; i <= len(s)-maxConsecutive; i++ {
		sub := s[i : i+maxConsecutive]
		
		// 检查连续相同字符
		if isAllSame(sub) {
			return true
		}

		// 检查连续数字或字母
		if isConsecutive(sub) {
			return true
		}
	}

	return false
}

// hasRepeatingSubstring 检查是否有重复子串
func (v *PasswordValidator) hasRepeatingSubstring(s string, minLength int) bool {
	if len(s) < minLength*2 {
		return false
	}

	for length := minLength; length <= len(s)/2; length++ {
		for i := 0; i <= len(s)-length*2; i++ {
			sub1 := s[i : i+length]
			sub2 := s[i+length : i+length*2]
			
			if sub1 == sub2 {
				return true
			}
		}
	}

	return false
}

// isAllSame 检查字符串中的所有字符是否相同
func isAllSame(s string) bool {
	if len(s) == 0 {
		return false
	}
	first := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != first {
			return false
		}
	}
	return true
}

// isConsecutive 检查字符串是否为连续字符
func isConsecutive(s string) bool {
	if len(s) < 2 {
		return false
	}

	// 检查数字序列
	if isNumericSequence(s) {
		return true
	}

	// 检查字母序列
	if isAlphaSequence(s) {
		return true
	}

	return false
}

// isNumericSequence 检查是否为数字序列
func isNumericSequence(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if !unicode.IsDigit(rune(s[i])) || !unicode.IsDigit(rune(s[i+1])) {
			return false
		}
		if abs(int(s[i+1])-int(s[i])) != 1 {
			return false
		}
	}
	return true
}

// isAlphaSequence 检查是否为字母序列
func isAlphaSequence(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if !unicode.IsLetter(rune(s[i])) || !unicode.IsLetter(rune(s[i+1])) {
			return false
		}
		if abs(int(unicode.ToLower(rune(s[i+1])))-int(unicode.ToLower(rune(s[i])))) != 1 {
			return false
		}
	}
	return true
}

// abs 绝对值函数
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// loadCommonPasswords 加载常见密码列表
func loadCommonPasswords() map[string]bool {
	common := []string{
		"password", "123456", "12345678", "123456789", "qwerty",
		"abc123", "password1", "12345", "1234567890", "111111",
		"1234567", "iloveyou", "admin", "welcome", "monkey",
		"letmein", "login", "abc123", "starwars", "123123",
		"dragon", "passw0rd", "master", "hello", "freedom",
		"whatever", "qazwsx", "trustno1", "654321", "jordan23",
		"harley", "password123", "1234", "robert", "matthew",
		"jordan", "asshole", "daniel", "andrew", "lakers",
		"andrea", "justin", "love", "jennifer", "sunshine",
		"buster", "123456", "computer", "amanda", "carlos",
		"dallas", "jessica", "pepper", "555555", "hannah",
		"thomas", "tigger", "robert", "soccer", "batman",
		"test", "pass", "killer", "hunter", "silver",
		"joseph", "michelle", "7777777", "diamond", "oliver",
		"mercedes", "benjamin", "samsung", "victoria", "jackson",
		"anthony", "joshua", "orange", "michelle", "london",
		"donald", "charles", "ginger", "sophie", "aaaaaa",
		"password1", "password123", "qwerty123", "admin123", "welcome123",
		"letmein123", "monkey123", "dragon123", "sunshine123", "master123",
		"hello123", "freedom123", "whatever123", "trustno1123", "jordan23123",
	}

	passwordMap := make(map[string]bool)
	for _, pwd := range common {
		passwordMap[pwd] = true
	}

	return passwordMap
}

// GenerateStrongPassword 生成强密码
func (v *PasswordValidator) GenerateStrongPassword() (string, error) {
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits    = "0123456789"
		specials = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	var charset string
	if v.RequireLower {
		charset += lowercase
	}
	if v.RequireUpper {
		charset += uppercase
	}
	if v.RequireNumber {
		charset += digits
	}
	if v.RequireSpecial {
		charset += specials
	}

	if charset == "" {
		charset = lowercase + uppercase + digits + specials
	}

	// 确保密码长度至少为12个字符
	length := v.MinLength
	if length < 12 {
		length = 12
	}

	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}

	// 验证生成的密码是否符合要求
	if err := v.ValidatePassword(string(password)); err != nil {
		// 如果生成的密码不符合要求，重新生成
		return v.GenerateStrongPassword()
	}

	return string(password), nil
}

// PasswordHistoryValidator 密码历史验证器
type PasswordHistoryValidator struct {
	HistorySize int
	History     []string
}

// NewPasswordHistoryValidator 创建密码历史验证器
func NewPasswordHistoryValidator(historySize int) *PasswordHistoryValidator {
	return &PasswordHistoryValidator{
		HistorySize: historySize,
		History:     make([]string, 0),
	}
}

// AddPassword 添加密码到历史记录
func (v *PasswordHistoryValidator) AddPassword(password string) {
	v.History = append(v.History, password)
	if len(v.History) > v.HistorySize {
		v.History = v.History[1:]
	}
}

// IsPasswordInHistory 检查密码是否在历史记录中
func (v *PasswordHistoryValidator) IsPasswordInHistory(password string) bool {
	for _, oldPassword := range v.History {
		if oldPassword == password {
			return true
		}
	}
	return false
}