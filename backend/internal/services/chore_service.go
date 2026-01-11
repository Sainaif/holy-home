package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/sainaif/holy-home/internal/models"
	"github.com/sainaif/holy-home/internal/repository"
)

type ChoreService struct {
	chores              repository.ChoreRepository
	choreAssignments    repository.ChoreAssignmentRepository
	choreSwapRequests   repository.ChoreSwapRequestRepository
	users               repository.UserRepository
	notificationService *NotificationService
}

func NewChoreService(
	chores repository.ChoreRepository,
	choreAssignments repository.ChoreAssignmentRepository,
	choreSwapRequests repository.ChoreSwapRequestRepository,
	users repository.UserRepository,
	notificationService *NotificationService,
) *ChoreService {
	return &ChoreService{
		chores:              chores,
		choreAssignments:    choreAssignments,
		choreSwapRequests:   choreSwapRequests,
		users:               users,
		notificationService: notificationService,
	}
}

type CreateChoreRequest struct {
	Name                 string  `json:"name"`
	Description          *string `json:"description,omitempty"`
	Frequency            string  `json:"frequency"` // daily, weekly, monthly, custom, irregular
	CustomInterval       *int    `json:"customInterval,omitempty"`
	Difficulty           int     `json:"difficulty"`     // 1-5
	Priority             int     `json:"priority"`       // 1-5
	AssignmentMode       string  `json:"assignmentMode"` // manual, round_robin, random
	NotificationsEnabled bool    `json:"notificationsEnabled"`
	ReminderHours        *int    `json:"reminderHours,omitempty"`
}

type AssignChoreRequest struct {
	ChoreID        string    `json:"choreId"`
	AssigneeUserID string    `json:"assigneeUserId"`
	DueDate        time.Time `json:"dueDate"`
}

type UpdateChoreAssignmentRequest struct {
	Status string `json:"status"` // pending, in_progress, done, overdue
}

type UpdateChoreRequest struct {
	Name                 *string `json:"name,omitempty"`
	Description          *string `json:"description,omitempty"`
	Frequency            *string `json:"frequency,omitempty"`
	CustomInterval       *int    `json:"customInterval,omitempty"`
	Difficulty           *int    `json:"difficulty,omitempty"`
	Priority             *int    `json:"priority,omitempty"`
	AssignmentMode       *string `json:"assignmentMode,omitempty"`
	ManualAssigneeId     *string `json:"manualAssigneeId,omitempty"` // User ID for manual assignment mode
	NotificationsEnabled *bool   `json:"notificationsEnabled,omitempty"`
	ReminderHours        *int    `json:"reminderHours,omitempty"`
	IsActive             *bool   `json:"isActive,omitempty"`
}

type ReassignChoreRequest struct {
	NewAssigneeUserID string `json:"newAssigneeUserId"`
}

type RandomAssignRequest struct {
	DueDate         time.Time `json:"dueDate"`
	EligibleUserIDs []string  `json:"eligibleUserIds"` // Users to randomly choose from
}

type CreateSwapRequest struct {
	RequesterAssignmentID string  `json:"requesterAssignmentId"`
	TargetAssignmentID    string  `json:"targetAssignmentId"`
	Message               *string `json:"message,omitempty"`
}

type RespondSwapRequest struct {
	ResponseMessage *string `json:"responseMessage,omitempty"`
}

type ChoreWithAssignment struct {
	Chore      models.Chore            `json:"chore"`
	Assignment *models.ChoreAssignment `json:"assignment,omitempty"`
}

// CreateChore creates a new chore (ADMIN only)
func (s *ChoreService) CreateChore(ctx context.Context, req CreateChoreRequest) (*models.Chore, error) {
	if req.Name == "" {
		return nil, errors.New("chore name is required")
	}

	// Set defaults
	if req.Frequency == "" {
		req.Frequency = "weekly"
	}
	if req.Difficulty < 1 {
		req.Difficulty = 1
	}
	if req.Priority < 1 {
		req.Priority = 1
	}
	if req.AssignmentMode == "" {
		req.AssignmentMode = "round_robin"
	}

	chore := models.Chore{
		ID:                   uuid.New().String(),
		Name:                 req.Name,
		Description:          req.Description,
		Frequency:            req.Frequency,
		CustomInterval:       req.CustomInterval,
		Difficulty:           req.Difficulty,
		Priority:             req.Priority,
		AssignmentMode:       req.AssignmentMode,
		NotificationsEnabled: req.NotificationsEnabled,
		ReminderHours:        req.ReminderHours,
		IsActive:             true,
		CreatedAt:            time.Now(),
	}

	if err := s.chores.Create(ctx, &chore); err != nil {
		return nil, fmt.Errorf("failed to create chore: %w", err)
	}

	log.Printf("[CHORE] Created: %q (ID: %s, frequency: %s, difficulty: %d)", chore.Name, chore.ID, chore.Frequency, chore.Difficulty)

	// Send notifications to all active users about the new chore
	if s.notificationService != nil {
		users, err := s.users.ListActive(ctx)
		if err == nil {
			now := time.Now()
			for _, user := range users {
				notification := &models.Notification{
					ID:           uuid.New().String(),
					UserID:       &user.ID,
					Channel:      "app",
					TemplateID:   "chore",
					ScheduledFor: now,
					SentAt:       &now,
					Status:       "sent",
					Title:        "Nowe zadanie domowe",
					Body:         fmt.Sprintf("Dodano nowe zadanie: %s", chore.Name),
				}
				s.notificationService.CreateNotification(ctx, notification)
			}
		}
	}

	return &chore, nil
}

// GetChores retrieves all chores
func (s *ChoreService) GetChores(ctx context.Context) ([]models.Chore, error) {
	chores, err := s.chores.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return chores, nil
}

// GetChore retrieves a chore by ID
func (s *ChoreService) GetChore(ctx context.Context, choreID string) (*models.Chore, error) {
	chore, err := s.chores.GetByID(ctx, choreID)
	if err != nil {
		return nil, errors.New("chore not found")
	}
	if chore == nil {
		return nil, errors.New("chore not found")
	}
	return chore, nil
}

// GetUserByID retrieves a user by ID
func (s *ChoreService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// AssignChore assigns a chore to a user (ADMIN only)
func (s *ChoreService) AssignChore(ctx context.Context, req AssignChoreRequest) (*models.ChoreAssignment, error) {
	// Verify chore exists and get it
	chore, err := s.GetChore(ctx, req.ChoreID)
	if err != nil {
		return nil, err
	}

	// Verify user exists
	_, err = s.users.GetByID(ctx, req.AssigneeUserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Calculate base points (difficulty * 10)
	points := chore.Difficulty * 10

	assignment := models.ChoreAssignment{
		ID:             uuid.New().String(),
		ChoreID:        req.ChoreID,
		AssigneeUserID: req.AssigneeUserID,
		DueDate:        req.DueDate,
		Status:         "pending",
		Points:         points,
		IsOnTime:       false,
	}

	if err := s.choreAssignments.Create(ctx, &assignment); err != nil {
		return nil, fmt.Errorf("failed to create chore assignment: %w", err)
	}

	log.Printf("[CHORE] Assigned: chore %q (ID: %s) to user %s, due: %s", chore.Name, req.ChoreID, req.AssigneeUserID, req.DueDate.Format("2006-01-02"))

	// Send notification to the assigned user
	if s.notificationService != nil {
		now := time.Now()
		notification := &models.Notification{
			ID:           uuid.New().String(),
			UserID:       &req.AssigneeUserID,
			Channel:      "app",
			TemplateID:   "chore",
			ScheduledFor: now,
			SentAt:       &now,
			Status:       "sent",
			Title:        "Przypisano zadanie",
			Body:         fmt.Sprintf("Przypisano Ci zadanie: %s (termin: %s)", chore.Name, req.DueDate.Format("2006-01-02")),
		}
		s.notificationService.CreateNotification(ctx, notification)
	}

	return &assignment, nil
}

// GetChoreAssignments retrieves all chore assignments
func (s *ChoreService) GetChoreAssignments(ctx context.Context, userID *string, status *string) ([]models.ChoreAssignment, error) {
	var assignments []models.ChoreAssignment
	var err error

	if userID != nil && status != nil {
		// Get by user then filter by status
		assignments, err = s.choreAssignments.ListByAssigneeID(ctx, *userID)
		if err != nil {
			return nil, fmt.Errorf("database error: %w", err)
		}
		// Filter by status
		filtered := []models.ChoreAssignment{}
		for _, a := range assignments {
			if a.Status == *status {
				filtered = append(filtered, a)
			}
		}
		assignments = filtered
	} else if userID != nil {
		assignments, err = s.choreAssignments.ListByAssigneeID(ctx, *userID)
	} else if status != nil {
		assignments, err = s.choreAssignments.ListByStatus(ctx, *status)
	} else {
		// List all - we'll use chores to get all assignments
		chores, err := s.chores.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("database error: %w", err)
		}
		for _, chore := range chores {
			choreAssignments, err := s.choreAssignments.ListByChoreID(ctx, chore.ID)
			if err != nil {
				continue
			}
			assignments = append(assignments, choreAssignments...)
		}
		return assignments, nil
	}

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return assignments, nil
}

// GetChoreAssignment retrieves a chore assignment by ID
func (s *ChoreService) GetChoreAssignment(ctx context.Context, assignmentID string) (*models.ChoreAssignment, error) {
	assignment, err := s.choreAssignments.GetByID(ctx, assignmentID)
	if err != nil {
		return nil, errors.New("chore assignment not found")
	}
	if assignment == nil {
		return nil, errors.New("chore assignment not found")
	}
	return assignment, nil
}

// UpdateChoreAssignment updates a chore assignment status
func (s *ChoreService) UpdateChoreAssignment(ctx context.Context, assignmentID string, req UpdateChoreAssignmentRequest) error {
	validStatuses := map[string]bool{
		"pending": true, "in_progress": true, "done": true, "overdue": true,
	}
	if !validStatuses[req.Status] {
		return errors.New("invalid status")
	}

	// Get current assignment
	assignment, err := s.GetChoreAssignment(ctx, assignmentID)
	if err != nil {
		return err
	}

	assignment.Status = req.Status

	if req.Status == "done" {
		now := time.Now()
		assignment.CompletedAt = &now

		// Check if completed on time and award bonus points
		isOnTime := now.Before(assignment.DueDate) || now.Equal(assignment.DueDate)
		assignment.IsOnTime = isOnTime

		if isOnTime {
			// 50% bonus for on-time completion
			assignment.Points = int(float64(assignment.Points) * 1.5)
		}
	} else {
		assignment.CompletedAt = nil
	}

	if err := s.choreAssignments.Update(ctx, assignment); err != nil {
		return fmt.Errorf("failed to update chore assignment: %w", err)
	}

	log.Printf("[CHORE] Assignment updated: ID=%s, status=%s, points=%d, on_time=%v", assignmentID, req.Status, assignment.Points, assignment.IsOnTime)

	return nil
}

// SwapChoreAssignment swaps two chore assignments (ADMIN only)
func (s *ChoreService) SwapChoreAssignment(ctx context.Context, assignment1ID, assignment2ID string) error {
	// Get both assignments
	assignment1, err := s.GetChoreAssignment(ctx, assignment1ID)
	if err != nil {
		return err
	}

	assignment2, err := s.GetChoreAssignment(ctx, assignment2ID)
	if err != nil {
		return err
	}

	// Swap assignees
	assignment1.AssigneeUserID, assignment2.AssigneeUserID = assignment2.AssigneeUserID, assignment1.AssigneeUserID

	if err := s.choreAssignments.Update(ctx, assignment1); err != nil {
		return fmt.Errorf("failed to update first assignment: %w", err)
	}

	if err := s.choreAssignments.Update(ctx, assignment2); err != nil {
		return fmt.Errorf("failed to update second assignment: %w", err)
	}

	return nil
}

// RotateChore creates a new assignment based on a rotating schedule (ADMIN only)
func (s *ChoreService) RotateChore(ctx context.Context, choreID string, dueDate time.Time) (*models.ChoreAssignment, error) {
	// Get all active users
	users, err := s.users.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if len(users) == 0 {
		return nil, errors.New("no active users to assign chore to")
	}

	// Get the last assignment for this chore
	lastAssignment, err := s.choreAssignments.GetLatestByChoreID(ctx, choreID)

	var nextUserID string

	if err != nil || lastAssignment == nil {
		// No previous assignment, assign to first user
		nextUserID = users[0].ID
	} else {
		// Find next user in rotation
		lastUserIndex := -1
		for i, u := range users {
			if u.ID == lastAssignment.AssigneeUserID {
				lastUserIndex = i
				break
			}
		}

		if lastUserIndex == -1 {
			// Last user no longer exists, start from beginning
			nextUserID = users[0].ID
		} else {
			// Move to next user (circular)
			nextUserIndex := (lastUserIndex + 1) % len(users)
			nextUserID = users[nextUserIndex].ID
		}
	}

	// Create new assignment
	return s.AssignChore(ctx, AssignChoreRequest{
		ChoreID:        choreID,
		AssigneeUserID: nextUserID,
		DueDate:        dueDate,
	})
}

// GetChoresWithAssignments retrieves chores with their current assignments
func (s *ChoreService) GetChoresWithAssignments(ctx context.Context) ([]ChoreWithAssignment, error) {
	chores, err := s.GetChores(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ChoreWithAssignment, 0, len(chores))

	for _, chore := range chores {
		// Get most recent pending assignment for this chore
		pendingAssignments, err := s.choreAssignments.ListPendingByAssignee(ctx, "")
		if err != nil {
			pendingAssignments = []models.ChoreAssignment{}
		}

		choreWithAssignment := ChoreWithAssignment{
			Chore: chore,
		}

		// Find pending assignment for this chore
		for _, assignment := range pendingAssignments {
			if assignment.ChoreID == chore.ID {
				a := assignment // Create a copy to avoid pointer issues
				choreWithAssignment.Assignment = &a
				break
			}
		}

		// If no pending, try to get latest assignment
		if choreWithAssignment.Assignment == nil {
			latest, err := s.choreAssignments.GetLatestByChoreID(ctx, chore.ID)
			if err == nil && latest != nil && latest.Status == "pending" {
				choreWithAssignment.Assignment = latest
			}
		}

		result = append(result, choreWithAssignment)
	}

	return result, nil
}

// UserStats represents user statistics for chores
type UserStats struct {
	UserID          string  `json:"userId"`
	UserName        string  `json:"userName"`
	TotalPoints     int     `json:"totalPoints"`
	CompletedChores int     `json:"completedChores"`
	OnTimeRate      float64 `json:"onTimeRate"`
	PendingChores   int     `json:"pendingChores"`
}

// AutoAssignChore automatically assigns a chore to the user with least workload
func (s *ChoreService) AutoAssignChore(ctx context.Context, choreID string, dueDate time.Time) (*models.ChoreAssignment, error) {
	// Get all active users
	users, err := s.users.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if len(users) == 0 {
		return nil, errors.New("no active users to assign chore to")
	}

	// Calculate workload for each user (pending chores + their difficulty)
	type userWorkload struct {
		UserID   string
		Workload int // Sum of difficulty points from pending chores
		Count    int // Number of pending chores
	}

	workloads := make([]userWorkload, 0, len(users))

	for _, user := range users {
		// Get user's pending assignments
		pendingAssignments, err := s.choreAssignments.ListPendingByAssignee(ctx, user.ID)
		if err != nil {
			pendingAssignments = []models.ChoreAssignment{}
		}

		totalWorkload := 0
		for _, assignment := range pendingAssignments {
			// Get the chore to find its difficulty
			assignedChore, err := s.GetChore(ctx, assignment.ChoreID)
			if err == nil {
				totalWorkload += assignedChore.Difficulty
			}
		}

		workloads = append(workloads, userWorkload{
			UserID:   user.ID,
			Workload: totalWorkload,
			Count:    len(pendingAssignments),
		})
	}

	// Find user with minimum workload (prioritize by difficulty sum, then by count)
	minWorkload := workloads[0]
	for _, wl := range workloads[1:] {
		if wl.Workload < minWorkload.Workload || (wl.Workload == minWorkload.Workload && wl.Count < minWorkload.Count) {
			minWorkload = wl
		}
	}

	// Assign to user with minimum workload
	return s.AssignChore(ctx, AssignChoreRequest{
		ChoreID:        choreID,
		AssigneeUserID: minWorkload.UserID,
		DueDate:        dueDate,
	})
}

// GetUserLeaderboard retrieves user rankings based on points
func (s *ChoreService) GetUserLeaderboard(ctx context.Context) ([]UserStats, error) {
	// Get all active users
	users, err := s.users.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	result := make([]UserStats, 0, len(users))

	for _, user := range users {
		// Get all completed assignments for this user
		allAssignments, err := s.choreAssignments.ListByAssigneeID(ctx, user.ID)
		if err != nil {
			continue
		}

		// Filter completed
		completedAssignments := []models.ChoreAssignment{}
		pendingCount := 0
		for _, a := range allAssignments {
			if a.Status == "done" {
				completedAssignments = append(completedAssignments, a)
			} else if a.Status == "pending" {
				pendingCount++
			}
		}

		// Calculate stats
		totalPoints := 0
		onTimeCount := 0
		for _, assignment := range completedAssignments {
			totalPoints += assignment.Points
			if assignment.IsOnTime {
				onTimeCount++
			}
		}

		onTimeRate := 0.0
		if len(completedAssignments) > 0 {
			onTimeRate = float64(onTimeCount) / float64(len(completedAssignments)) * 100
		}

		result = append(result, UserStats{
			UserID:          user.ID,
			UserName:        user.Name,
			TotalPoints:     totalPoints,
			CompletedChores: len(completedAssignments),
			OnTimeRate:      onTimeRate,
			PendingChores:   pendingCount,
		})
	}

	// Sort by total points descending
	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if result[j].TotalPoints < result[j+1].TotalPoints {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return result, nil
}

// DeleteChore deletes a chore and all its assignments
func (s *ChoreService) DeleteChore(ctx context.Context, choreID string) error {
	// Delete all assignments for this chore
	assignments, err := s.choreAssignments.ListByChoreID(ctx, choreID)
	if err != nil {
		return fmt.Errorf("failed to list chore assignments: %w", err)
	}

	for _, assignment := range assignments {
		if err := s.choreAssignments.Delete(ctx, assignment.ID); err != nil {
			return fmt.Errorf("failed to delete chore assignment: %w", err)
		}
	}

	// Delete the chore
	if err := s.chores.Delete(ctx, choreID); err != nil {
		return fmt.Errorf("failed to delete chore: %w", err)
	}

	log.Printf("[CHORE] Deleted: ID=%s (including %d assignments)", choreID, len(assignments))

	return nil
}

// UpdateChore updates an existing chore
func (s *ChoreService) UpdateChore(ctx context.Context, choreID string, req UpdateChoreRequest) (*models.Chore, error) {
	// Get existing chore
	chore, err := s.GetChore(ctx, choreID)
	if err != nil {
		return nil, err
	}

	// Apply updates (only non-nil fields)
	if req.Name != nil {
		if *req.Name == "" {
			return nil, errors.New("chore name cannot be empty")
		}
		chore.Name = *req.Name
	}
	if req.Description != nil {
		chore.Description = req.Description
	}
	if req.Frequency != nil {
		validFrequencies := map[string]bool{
			"daily": true, "weekly": true, "monthly": true, "custom": true, "irregular": true,
		}
		if !validFrequencies[*req.Frequency] {
			return nil, errors.New("invalid frequency")
		}
		chore.Frequency = *req.Frequency
	}
	if req.CustomInterval != nil {
		chore.CustomInterval = req.CustomInterval
	}
	if req.Difficulty != nil {
		if *req.Difficulty < 1 || *req.Difficulty > 5 {
			return nil, errors.New("difficulty must be between 1 and 5")
		}
		chore.Difficulty = *req.Difficulty
	}
	if req.Priority != nil {
		if *req.Priority < 1 || *req.Priority > 5 {
			return nil, errors.New("priority must be between 1 and 5")
		}
		chore.Priority = *req.Priority
	}
	if req.AssignmentMode != nil {
		validModes := map[string]bool{
			"manual": true, "round_robin": true, "random": true,
		}
		if !validModes[*req.AssignmentMode] {
			return nil, errors.New("invalid assignment mode")
		}
		chore.AssignmentMode = *req.AssignmentMode
	}
	if req.NotificationsEnabled != nil {
		chore.NotificationsEnabled = *req.NotificationsEnabled
	}
	if req.ReminderHours != nil {
		chore.ReminderHours = req.ReminderHours
	}
	if req.IsActive != nil {
		chore.IsActive = *req.IsActive
	}

	if err := s.chores.Update(ctx, chore); err != nil {
		return nil, fmt.Errorf("failed to update chore: %w", err)
	}

	// Handle manual assignment if provided
	if req.ManualAssigneeId != nil && *req.ManualAssigneeId != "" && chore.AssignmentMode == "manual" {
		// Verify user exists
		user, err := s.users.GetByID(ctx, *req.ManualAssigneeId)
		if err != nil || user == nil {
			return nil, errors.New("manual assignee user not found")
		}

		// Get current pending/in_progress assignments for this chore
		assignments, err := s.choreAssignments.ListByChoreID(ctx, choreID)
		if err != nil {
			log.Printf("[CHORE] Warning: failed to get assignments for chore %s: %v", choreID, err)
		} else {
			// Reassign pending/in_progress assignments to the selected user
			for _, assignment := range assignments {
				if assignment.Status == "pending" || assignment.Status == "in_progress" {
					assignment.AssigneeUserID = *req.ManualAssigneeId
					if err := s.choreAssignments.Update(ctx, &assignment); err != nil {
						log.Printf("[CHORE] Warning: failed to reassign assignment %s: %v", assignment.ID, err)
					} else {
						log.Printf("[CHORE] Reassigned assignment %s to user %s", assignment.ID, *req.ManualAssigneeId)
					}
				}
			}
		}
	}

	log.Printf("[CHORE] Updated: %q (ID: %s)", chore.Name, chore.ID)

	return chore, nil
}

// ReassignChoreAssignment reassigns a chore to a different user
func (s *ChoreService) ReassignChoreAssignment(ctx context.Context, assignmentID string, req ReassignChoreRequest) (*models.ChoreAssignment, error) {
	if req.NewAssigneeUserID == "" {
		return nil, errors.New("new assignee user ID is required")
	}

	// Verify new user exists
	newUser, err := s.users.GetByID(ctx, req.NewAssigneeUserID)
	if err != nil || newUser == nil {
		return nil, errors.New("new assignee user not found")
	}

	// Get current assignment
	assignment, err := s.GetChoreAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	// Only allow reassigning pending or in_progress assignments
	if assignment.Status != "pending" && assignment.Status != "in_progress" {
		return nil, errors.New("can only reassign pending or in-progress chores")
	}

	oldAssigneeID := assignment.AssigneeUserID
	assignment.AssigneeUserID = req.NewAssigneeUserID

	if err := s.choreAssignments.Update(ctx, assignment); err != nil {
		return nil, fmt.Errorf("failed to reassign chore: %w", err)
	}

	log.Printf("[CHORE] Reassigned: assignment %s from user %s to user %s", assignmentID, oldAssigneeID, req.NewAssigneeUserID)

	// Send notification to new assignee
	if s.notificationService != nil {
		chore, _ := s.GetChore(ctx, assignment.ChoreID)
		if chore != nil {
			now := time.Now()
			notification := &models.Notification{
				ID:           uuid.New().String(),
				UserID:       &req.NewAssigneeUserID,
				Channel:      "app",
				TemplateID:   "chore",
				ScheduledFor: now,
				SentAt:       &now,
				Status:       "sent",
				Title:        "Przypisano zadanie",
				Body:         fmt.Sprintf("Przypisano Ci zadanie: %s (termin: %s)", chore.Name, assignment.DueDate.Format("2006-01-02")),
			}
			s.notificationService.CreateNotification(ctx, notification)
		}
	}

	return assignment, nil
}

// RandomAssignChore randomly assigns a chore to one of the eligible users
func (s *ChoreService) RandomAssignChore(ctx context.Context, choreID string, req RandomAssignRequest) (*models.ChoreAssignment, error) {
	// Verify chore exists
	_, err := s.GetChore(ctx, choreID)
	if err != nil {
		return nil, err
	}

	// If no eligible users provided, use all active users
	eligibleUserIDs := req.EligibleUserIDs
	if len(eligibleUserIDs) == 0 {
		users, err := s.users.ListActive(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get active users: %w", err)
		}
		for _, u := range users {
			eligibleUserIDs = append(eligibleUserIDs, u.ID)
		}
	}

	if len(eligibleUserIDs) == 0 {
		return nil, errors.New("no eligible users to assign chore to")
	}

	// Verify all eligible users exist
	for _, userID := range eligibleUserIDs {
		user, err := s.users.GetByID(ctx, userID)
		if err != nil || user == nil {
			return nil, fmt.Errorf("user %s not found", userID)
		}
	}

	// Pick a random user using crypto/rand for secure randomness
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(eligibleUserIDs))))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %w", err)
	}
	randomUserID := eligibleUserIDs[n.Int64()]

	log.Printf("[CHORE] Random assignment: selected user %s from %d eligible users", randomUserID, len(eligibleUserIDs))

	// Assign to random user
	return s.AssignChore(ctx, AssignChoreRequest{
		ChoreID:        choreID,
		AssigneeUserID: randomUserID,
		DueDate:        req.DueDate,
	})
}

// ============================================
// SWAP REQUEST METHODS (User-to-User Swap Approval)
// ============================================

// CreateSwapRequest creates a new swap request from one user to another
func (s *ChoreService) CreateSwapRequest(ctx context.Context, requesterUserID string, req CreateSwapRequest) (*models.ChoreSwapRequest, error) {
	// Get requester's assignment
	requesterAssignment, err := s.GetChoreAssignment(ctx, req.RequesterAssignmentID)
	if err != nil {
		return nil, errors.New("requester assignment not found")
	}

	// Verify requester owns this assignment
	if requesterAssignment.AssigneeUserID != requesterUserID {
		return nil, errors.New("you can only swap your own assignments")
	}

	// Get target assignment
	targetAssignment, err := s.GetChoreAssignment(ctx, req.TargetAssignmentID)
	if err != nil {
		return nil, errors.New("target assignment not found")
	}

	// Prevent swapping with yourself
	if targetAssignment.AssigneeUserID == requesterUserID {
		return nil, errors.New("cannot swap with yourself")
	}

	// Both must be pending or in_progress
	if requesterAssignment.Status != "pending" && requesterAssignment.Status != "in_progress" {
		return nil, errors.New("requester assignment must be pending or in progress")
	}
	if targetAssignment.Status != "pending" && targetAssignment.Status != "in_progress" {
		return nil, errors.New("target assignment must be pending or in progress")
	}

	// Check for existing pending swap request involving these assignments
	existingRequests, err := s.choreSwapRequests.ListPendingForAssignment(ctx, req.RequesterAssignmentID)
	if err == nil && len(existingRequests) > 0 {
		return nil, errors.New("there is already a pending swap request for this assignment")
	}
	existingRequests, err = s.choreSwapRequests.ListPendingForAssignment(ctx, req.TargetAssignmentID)
	if err == nil && len(existingRequests) > 0 {
		return nil, errors.New("target assignment already has a pending swap request")
	}

	// Create swap request with 48 hour expiration
	now := time.Now()
	swapRequest := &models.ChoreSwapRequest{
		ID:                    uuid.New().String(),
		RequesterUserID:       requesterUserID,
		RequesterAssignmentID: req.RequesterAssignmentID,
		TargetUserID:          targetAssignment.AssigneeUserID,
		TargetAssignmentID:    req.TargetAssignmentID,
		Status:                "pending",
		Message:               req.Message,
		ExpiresAt:             now.Add(48 * time.Hour),
		CreatedAt:             now,
	}

	if err := s.choreSwapRequests.Create(ctx, swapRequest); err != nil {
		return nil, fmt.Errorf("failed to create swap request: %w", err)
	}

	log.Printf("[CHORE] Swap request created: ID=%s, requester=%s, target=%s", swapRequest.ID, requesterUserID, targetAssignment.AssigneeUserID)

	// Send notification to target user
	if s.notificationService != nil {
		requester, _ := s.users.GetByID(ctx, requesterUserID)
		requesterChore, _ := s.GetChore(ctx, requesterAssignment.ChoreID)
		targetChore, _ := s.GetChore(ctx, targetAssignment.ChoreID)

		if requester != nil && requesterChore != nil && targetChore != nil {
			notification := &models.Notification{
				ID:           uuid.New().String(),
				UserID:       &targetAssignment.AssigneeUserID,
				Channel:      "app",
				TemplateID:   "chore_swap_request",
				ScheduledFor: now,
				SentAt:       &now,
				Status:       "sent",
				Title:        "Prośba o zamianę zadania",
				Body:         fmt.Sprintf("%s chce zamienić zadanie '%s' na Twoje '%s'", requester.Name, requesterChore.Name, targetChore.Name),
			}
			s.notificationService.CreateNotification(ctx, notification)
		}
	}

	return swapRequest, nil
}

// AcceptSwapRequest accepts a swap request and performs the swap
func (s *ChoreService) AcceptSwapRequest(ctx context.Context, targetUserID, requestID string, req RespondSwapRequest) error {
	// Expire old requests first
	s.choreSwapRequests.ExpireOldRequests(ctx)

	swapRequest, err := s.choreSwapRequests.GetByID(ctx, requestID)
	if err != nil || swapRequest == nil {
		return errors.New("swap request not found")
	}

	// Verify target user
	if swapRequest.TargetUserID != targetUserID {
		return errors.New("you are not the target of this swap request")
	}

	// Check status
	if swapRequest.Status != "pending" {
		return fmt.Errorf("swap request is already %s", swapRequest.Status)
	}

	// Check expiration
	if time.Now().After(swapRequest.ExpiresAt) {
		swapRequest.Status = "expired"
		s.choreSwapRequests.Update(ctx, swapRequest)
		return errors.New("swap request has expired")
	}

	// Perform the actual swap
	if err := s.SwapChoreAssignment(ctx, swapRequest.RequesterAssignmentID, swapRequest.TargetAssignmentID); err != nil {
		return fmt.Errorf("failed to perform swap: %w", err)
	}

	// Update swap request
	now := time.Now()
	swapRequest.Status = "accepted"
	swapRequest.ResponseMessage = req.ResponseMessage
	swapRequest.RespondedAt = &now

	if err := s.choreSwapRequests.Update(ctx, swapRequest); err != nil {
		return fmt.Errorf("failed to update swap request: %w", err)
	}

	log.Printf("[CHORE] Swap request accepted: ID=%s", requestID)

	// Send notification to requester
	if s.notificationService != nil {
		target, _ := s.users.GetByID(ctx, targetUserID)
		if target != nil {
			notification := &models.Notification{
				ID:           uuid.New().String(),
				UserID:       &swapRequest.RequesterUserID,
				Channel:      "app",
				TemplateID:   "chore_swap_accepted",
				ScheduledFor: now,
				SentAt:       &now,
				Status:       "sent",
				Title:        "Zamiana zaakceptowana",
				Body:         fmt.Sprintf("%s zaakceptował(a) Twoją prośbę o zamianę zadań", target.Name),
			}
			s.notificationService.CreateNotification(ctx, notification)
		}
	}

	return nil
}

// RejectSwapRequest rejects a swap request
func (s *ChoreService) RejectSwapRequest(ctx context.Context, targetUserID, requestID string, req RespondSwapRequest) error {
	// Expire old requests first
	s.choreSwapRequests.ExpireOldRequests(ctx)

	swapRequest, err := s.choreSwapRequests.GetByID(ctx, requestID)
	if err != nil || swapRequest == nil {
		return errors.New("swap request not found")
	}

	// Verify target user
	if swapRequest.TargetUserID != targetUserID {
		return errors.New("you are not the target of this swap request")
	}

	// Check status
	if swapRequest.Status != "pending" {
		return fmt.Errorf("swap request is already %s", swapRequest.Status)
	}

	// Update swap request
	now := time.Now()
	swapRequest.Status = "rejected"
	swapRequest.ResponseMessage = req.ResponseMessage
	swapRequest.RespondedAt = &now

	if err := s.choreSwapRequests.Update(ctx, swapRequest); err != nil {
		return fmt.Errorf("failed to update swap request: %w", err)
	}

	log.Printf("[CHORE] Swap request rejected: ID=%s", requestID)

	// Send notification to requester
	if s.notificationService != nil {
		target, _ := s.users.GetByID(ctx, targetUserID)
		if target != nil {
			notification := &models.Notification{
				ID:           uuid.New().String(),
				UserID:       &swapRequest.RequesterUserID,
				Channel:      "app",
				TemplateID:   "chore_swap_rejected",
				ScheduledFor: now,
				SentAt:       &now,
				Status:       "sent",
				Title:        "Zamiana odrzucona",
				Body:         fmt.Sprintf("%s odrzucił(a) Twoją prośbę o zamianę zadań", target.Name),
			}
			s.notificationService.CreateNotification(ctx, notification)
		}
	}

	return nil
}

// CancelSwapRequest cancels a pending swap request by the requester
func (s *ChoreService) CancelSwapRequest(ctx context.Context, requesterUserID, requestID string) error {
	swapRequest, err := s.choreSwapRequests.GetByID(ctx, requestID)
	if err != nil || swapRequest == nil {
		return errors.New("swap request not found")
	}

	// Verify requester
	if swapRequest.RequesterUserID != requesterUserID {
		return errors.New("you can only cancel your own swap requests")
	}

	// Check status
	if swapRequest.Status != "pending" {
		return fmt.Errorf("swap request is already %s", swapRequest.Status)
	}

	// Update swap request
	now := time.Now()
	swapRequest.Status = "cancelled"
	swapRequest.RespondedAt = &now

	if err := s.choreSwapRequests.Update(ctx, swapRequest); err != nil {
		return fmt.Errorf("failed to cancel swap request: %w", err)
	}

	log.Printf("[CHORE] Swap request cancelled: ID=%s", requestID)

	return nil
}

// GetPendingSwapRequestsForUser gets all pending swap requests where user is the target
func (s *ChoreService) GetPendingSwapRequestsForUser(ctx context.Context, userID string) ([]models.ChoreSwapRequest, error) {
	// Expire old requests first
	s.choreSwapRequests.ExpireOldRequests(ctx)

	requests, err := s.choreSwapRequests.ListPendingByTargetID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap requests: %w", err)
	}
	return requests, nil
}

// GetMySwapRequests gets all swap requests created by the user
func (s *ChoreService) GetMySwapRequests(ctx context.Context, userID string) ([]models.ChoreSwapRequest, error) {
	// Expire old requests first
	s.choreSwapRequests.ExpireOldRequests(ctx)

	requests, err := s.choreSwapRequests.ListByRequesterID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap requests: %w", err)
	}
	return requests, nil
}

// GetSwapRequest gets a swap request by ID
func (s *ChoreService) GetSwapRequest(ctx context.Context, requestID string) (*models.ChoreSwapRequest, error) {
	request, err := s.choreSwapRequests.GetByID(ctx, requestID)
	if err != nil || request == nil {
		return nil, errors.New("swap request not found")
	}
	return request, nil
}
