package service

import (
	"context"

	"google.golang.org/api/idtoken"
)

type GoogleIDTokenVerifier struct {
	clientID string
}

func NewGoogleIDTokenVerifier(clientID string) *GoogleIDTokenVerifier {
	return &GoogleIDTokenVerifier{clientID: clientID}
}

func (v *GoogleIDTokenVerifier) Verify(ctx context.Context, idToken string) (*GoogleClaims, error) {
	payload, err := idtoken.Validate(ctx, idToken, v.clientID)
	if err != nil {
		return nil, err
	}

	claims := &GoogleClaims{Subject: payload.Subject}
	if email, ok := payload.Claims["email"].(string); ok {
		claims.Email = email
	}
	if verified, ok := payload.Claims["email_verified"].(bool); ok {
		claims.EmailVerified = verified
	}
	if given, ok := payload.Claims["given_name"].(string); ok {
		claims.GivenName = given
	}
	if family, ok := payload.Claims["family_name"].(string); ok {
		claims.FamilyName = family
	}

	return claims, nil
}
