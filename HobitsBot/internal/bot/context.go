package bot

import (
	"sync"
	"time"
)

type UserContext struct {
	State       string            // Текущее состояние пользователя
	Data        map[string]string // Данные для текущего процесса
	MessageID   int               // ID последнего сообщения для удаления
	ChatID      int64             // Chat ID для удаления сообщений
	CreatedAt   time.Time
	LastUpdated time.Time
}

type ContextManager struct {
	contexts map[int64]*UserContext
	mu       sync.RWMutex
	ttl      time.Duration
}

// NewContextManager создает новый менеджер контекстов
func NewContextManager(ttl time.Duration) *ContextManager {
	cm := &ContextManager{
		contexts: make(map[int64]*UserContext),
		ttl:      ttl,
	}

	// Запускаем горутину для очистки устаревших контекстов
	go cm.cleanupOldContexts()

	return cm
}

// SetState устанавливает состояние пользователя
func (cm *ContextManager) SetState(userID int64, state string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ctx, ok := cm.contexts[userID]; ok {
		ctx.State = state
		ctx.LastUpdated = time.Now()
	} else {
		cm.contexts[userID] = &UserContext{
			State:       state,
			Data:        make(map[string]string),
			CreatedAt:   time.Now(),
			LastUpdated: time.Now(),
		}
	}
}

// GetState получает состояние пользователя
func (cm *ContextManager) GetState(userID int64) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if ctx, ok := cm.contexts[userID]; ok {
		return ctx.State
	}
	return ""
}

// SetData устанавливает данные пользователя
func (cm *ContextManager) SetData(userID int64, key, value string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ctx, ok := cm.contexts[userID]; ok {
		ctx.Data[key] = value
		ctx.LastUpdated = time.Now()
	} else {
		cm.contexts[userID] = &UserContext{
			State:       "",
			Data:        map[string]string{key: value},
			CreatedAt:   time.Now(),
			LastUpdated: time.Now(),
		}
	}
}

// GetData получает данные пользователя
func (cm *ContextManager) GetData(userID int64, key string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if ctx, ok := cm.contexts[userID]; ok {
		return ctx.Data[key]
	}
	return ""
}

// GetAllData получает все данные пользователя
func (cm *ContextManager) GetAllData(userID int64) map[string]string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if ctx, ok := cm.contexts[userID]; ok {
		data := make(map[string]string)
		for k, v := range ctx.Data {
			data[k] = v
		}
		return data
	}
	return make(map[string]string)
}

// ClearContext очищает контекст пользователя
func (cm *ContextManager) ClearContext(userID int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.contexts, userID)
}

// ClearData очищает конкретные данные пользователя
func (cm *ContextManager) ClearData(userID int64, key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ctx, ok := cm.contexts[userID]; ok {
		delete(ctx.Data, key)
		ctx.LastUpdated = time.Now()
	}
}

// HasKey проверяет наличие ключа в данных пользователя
func (cm *ContextManager) HasKey(userID int64, key string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if ctx, ok := cm.contexts[userID]; ok {
		_, exists := ctx.Data[key]
		return exists
	}
	return false
}

// SetMessageID сохраняет ID сообщения для удаления при переходе
func (cm *ContextManager) SetMessageID(userID int64, chatID int64, messageID int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ctx, ok := cm.contexts[userID]; ok {
		ctx.MessageID = messageID
		ctx.ChatID = chatID
		ctx.LastUpdated = time.Now()
	} else {
		cm.contexts[userID] = &UserContext{
			State:       "",
			Data:        make(map[string]string),
			MessageID:   messageID,
			ChatID:      chatID,
			CreatedAt:   time.Now(),
			LastUpdated: time.Now(),
		}
	}
}

// GetMessageID получает ID сохраненного сообщения
func (cm *ContextManager) GetMessageID(userID int64) (int64, int) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if ctx, ok := cm.contexts[userID]; ok {
		return ctx.ChatID, ctx.MessageID
	}
	return 0, 0
}

// ClearMessageID очищает сохраненный ID сообщения
func (cm *ContextManager) ClearMessageID(userID int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ctx, ok := cm.contexts[userID]; ok {
		ctx.MessageID = 0
		ctx.ChatID = 0
		ctx.LastUpdated = time.Now()
	}
}

// cleanupOldContexts периодически удаляет устаревшие контексты
func (cm *ContextManager) cleanupOldContexts() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cm.mu.Lock()
		now := time.Now()
		for userID, ctx := range cm.contexts {
			if now.Sub(ctx.LastUpdated) > cm.ttl {
				delete(cm.contexts, userID)
			}
		}
		cm.mu.Unlock()
	}
}
