package main

import "time"

type Where struct {
	Operator  string `json:"operator"`
	Connector string `json:"connector"`
	Field     string `json:"field"`
	Value     string `json:"value"`
}

type FindOneRequestBody struct {
	Model string  `json:"model"`
	Where []Where `json:"where"`
}

type CountRequestBody struct {
	Model string  `json:"model"`
	Where []Where `json:"where"`
}

type CreateRequestBody struct {
	Model  string      `json:"model"`
	Data   interface{} `json:"data"`
	Select []string    `json:"select"`
}

type DeleteRequestBody struct {
	Model string  `json:"model"`
	Where []Where `json:"where"`
}

type DeleteManyRequestBody struct {
	Model string  `json:"model"`
	Where []Where `json:"where"`
}

type FindManyRequestBody struct {
	Model  string   `json:"model"`
	Where  []Where  `json:"where"`
	Limit  int      `json:"limit"`
	SortBy []string `json:"sortBy"`
	Offset int      `json:"offset"`
}

type UpdateRequestBody struct {
	Model  string      `json:"model"`
	Where  []Where     `json:"where"`
	Update interface{} `json:"update"`
}

type UpdateManyRequestBody struct {
	Model  string      `json:"model"`
	Where  []Where     `json:"where"`
	Update interface{} `json:"update"`
}

type CreateSchemaRequestBody struct {
	// Define the structure for the create schema request body
}

// User represents the user table
type User struct {
	ID            string    `db:"id" pk:"true"`
	Name          string    `db:"name"`
	Email         string    `db:"email"`
	EmailVerified bool      `db:"emailVerified"`
	Image         string    `db:"image"`
	CreatedAt     time.Time `db:"createdAt"`
	UpdatedAt     time.Time `db:"updatedAt"`
}

// Session represents the session table
type Session struct {
	ID        string    `db:"id" pk:"true"`
	UserID    string    `db:"userId"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expiresAt"`
	IPAddress string    `db:"ipAddress"`
	UserAgent string    `db:"userAgent"`
	CreatedAt time.Time `db:"createdAt"`
	UpdatedAt time.Time `db:"updatedAt"`
}

// Account represents the account table
type Account struct {
	ID                   string    `db:"id" pk:"true"`
	UserID               string    `db:"userId"`
	AccountID            string    `db:"accountId"`
	ProviderID           string    `db:"providerId"`
	AccessToken          string    `db:"accessToken"`
	RefreshToken         string    `db:"refreshToken"`
	AccessTokenExpiresAt time.Time `db:"accessTokenExpiresAt"`
	RefreshTokenExpiresAt time.Time `db:"refreshTokenExpiresAt"`
	Scope                string    `db:"scope"`
	IDToken              string    `db:"idToken"`
	Password             string    `db:"password"`
	CreatedAt            time.Time `db:"createdAt"`
	UpdatedAt            time.Time `db:"updatedAt"`
}

// Verification represents the verification table
type Verification struct {
	ID         string    `db:"id" pk:"true"`
	Identifier string    `db:"identifier"`
	Value      string    `db:"value"`
	ExpiresAt  time.Time `db:"expiresAt"`
	CreatedAt  time.Time `db:"createdAt"`
	UpdatedAt  time.Time `db:"updatedAt"`
}
