// Package telegram provides Telegram Bot API integration.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "https://api.telegram.org/bot"

// Client is a low-level Telegram Bot API HTTP client.
type Client struct {
	token  string
	client *http.Client
}

// NewClient creates a new Telegram API client.
func NewClient(token string) *Client {
	return &Client{
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewClientWithLongPoll creates a client with a longer timeout for getUpdates.
func NewClientWithLongPoll(token string, pollTimeout time.Duration) *Client {
	return &Client{
		token: token,
		client: &http.Client{
			Timeout: pollTimeout + 5*time.Second,
		},
	}
}

// SendMessage sends a text message to a chat.
func (c *Client) SendMessage(ctx context.Context, chatID, text, parseMode string) error {
	payload := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": parseMode,
	}
	return c.post(ctx, "sendMessage", payload)
}

// SendMessageWithKeyboard sends a message with inline keyboard.
func (c *Client) SendMessageWithKeyboard(ctx context.Context, chatID, text, parseMode string, keyboard [][]InlineKeyboardButton) error {
	payload := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": parseMode,
		"reply_markup": map[string]any{
			"inline_keyboard": keyboard,
		},
	}
	return c.post(ctx, "sendMessage", payload)
}

// AnswerCallbackQuery answers a callback query.
func (c *Client) AnswerCallbackQuery(ctx context.Context, callbackQueryID, text string) error {
	payload := map[string]any{
		"callback_query_id": callbackQueryID,
		"text":              text,
	}
	return c.post(ctx, "answerCallbackQuery", payload)
}

// GetUpdates retrieves updates using long polling.
func (c *Client) GetUpdates(ctx context.Context, offset int, timeout int) ([]Update, error) {
	payload := map[string]any{
		"offset":  offset,
		"timeout": timeout,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := apiBase + c.token + "/getUpdates"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("telegram getUpdates parse error: %w", err)
	}
	if !result.OK {
		return nil, fmt.Errorf("telegram getUpdates failed: %s", string(respBody))
	}

	return result.Result, nil
}

func (c *Client) post(ctx context.Context, method string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := apiBase + c.token + "/" + method
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram %s failed (status %d): %s", method, resp.StatusCode, string(respBody))
	}

	return nil
}

// Update represents a Telegram update.
type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

// Message represents a Telegram message.
type Message struct {
	MessageID int    `json:"message_id"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
	From      *User  `json:"from,omitempty"`
}

// Chat represents a Telegram chat.
type Chat struct {
	ID int64 `json:"id"`
}

// User represents a Telegram user.
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"`
}

// CallbackQuery represents a Telegram callback query.
type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data"`
}

// InlineKeyboardButton represents a button in an inline keyboard.
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}
