package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

type CreateProfileResp struct {
	User User `json:"user"`
}

type ProfileResp struct {
	User UserDB `json:"user"`
}

// getUserProfile fetches the user's private profile info (and creates
// it if it doesn't exist)
func (h *Handler) getUserProfile(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		resp      = &ProfileResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	resp.User.Email = userEmail

	// Fetch the user
	iter := h.database.Collection("users").Where("email", "==", userEmail).Documents(ctx)
	for {
		doc, err := iter.Next()

		if doc == nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		var u UserDB
		doc.DataTo(&u)
		resp.User = u

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		break
	}
	json.NewEncoder(w).Encode(resp)
}

// createProfile initializes the profile for the user.
// The initial payload comes from Auth0 and has a default nickname
// and profile photo.
func (h *Handler) createProfile(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		err       error
		req       UserDB
		resp      = &CreateProfileResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "error handling reques", http.StatusBadRequest)
		return
	}

	// Validate user
	if req.Email != userEmail {
		http.Error(w, "only the user can create their profile", http.StatusBadRequest)
		return
	}

	// Fetch the user first to make sure it doesn't exist
	iter := h.database.Collection("users").Where("email", "==", userEmail).Documents(ctx)
	for {
		doc, err := iter.Next()

		if doc != nil && err != nil {
			http.Error(w, "user already exists", http.StatusBadRequest)
			return
		}

		// Create a new user record if it doesn't exist
		newUser := UserDB{
			Email:    req.Email,
			Username: req.Username,
			Photo:    req.Photo,
		}
		created, _, err := h.database.Collection("users").Add(ctx, newUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infow(
			"User added",
			"id", created.ID,
			"updated_by", userEmail,
		)

		resp.User = User{
			Photo:    newUser.Photo,
			Username: newUser.Username,
		}

		break
	}

	json.NewEncoder(w).Encode(resp)
}

// func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
// 	var (
// 		ctx       = context.TODO()
// 		docID     string
// 		err       error
// 		req       UserDB
// 		resp      = &CreateProfileResp{}
// 		userEmail = r.Header.Get("X-User-Email")
// 	)

// 	err = json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// Fetch the user
// 	iter := h.database.Collection("users").Where("email", "==", userEmail).Documents(ctx)
// 	for {
// 		doc, err := iter.Next()

// 		if doc == nil {
// 			http.Error(w, "invalid user", http.StatusBadRequest)
// 			return
// 		}

// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
// 		}

// 		docID = doc.Ref.ID

// 		break
// 	}

// 	// Update the user
// 	user := h.database.Collection("users").Doc(docID)
// 	docsnap, err := user.Get(ctx)
// 	if err != nil {
// 		h.logger.Error(err)
// 		http.Error(w, "invalid user", http.StatusBadRequest)
// 		return
// 	}

// 	result, _ := user.Update(
// 		ctx,
// 		[]firestore.Update{
// 			{Path: "email", Value: userEmail},
// 			{Path: "username", Value: req.Username},
// 		},
// 	)
// 	h.logger.Infow(
// 		"User updated",
// 		"id", docsnap.Ref.ID,
// 		"updated_at", result.UpdateTime,
// 		"updated_by", userEmail,
// 	)

// 	resp.User = UserDB{
// 		Username: req.Username,
// 		Email:    userEmail,
// 	}

// 	json.NewEncoder(w).Encode(resp)
// }
