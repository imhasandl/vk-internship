package subpub

import (
    "context"
    "sync"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestSubscribe проверяет функционал подписки на сообщения
func TestSubscribe(t *testing.T) {
    tests := []struct {
        name      string
        key       string
        wantErr   bool
    }{
        {
            name:      "Базовая подписка",
            key:       "topic1",
            wantErr:   false,
        },
        {
            name:      "Подписка с пустым ключом",
            key:       "",
            wantErr:   false,
        },
        {
            name:      "Подписка на ключ со специальными символами",
            key:       "test-key!@#$%^&*()",
            wantErr:   false,
        },
        {
            name:      "Русский ключ",
            key:       "тестовый-ключ",
            wantErr:   false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            pubSub := NewSubPub()
            
            // Создаем обработчик сообщений
            handler := func(msg interface{}) {}
            
            // Подписываемся на топик
            subscription, err := pubSub.Subscribe(tt.key, handler)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, subscription)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, subscription)
            }
            
            // Проверяем что подписка существует в структурах данных
            if !tt.wantErr {
                pubSub.mu.Lock()
                subscribers, ok := pubSub.subscribers[tt.key]
                pubSub.mu.Unlock()
                
                assert.True(t, ok, "Должна быть создана карта для ключа")
                assert.Equal(t, 1, len(subscribers), "Должен быть один подписчик")
            }
        })
    }
}

// TestPublish проверяет функционал публикации сообщений
func TestPublish(t *testing.T) {
    tests := []struct {
        name         string
        subscribeKey string
        publishKey   string
        message      string
        expectMsg    bool
    }{
        {
            name:         "Сообщение доставлено подписчику",
            subscribeKey: "test-topic",
            publishKey:   "test-topic",
            message:      "Hello world!",
            expectMsg:    true,
        },
        {
            name:         "Разные ключи - сообщение не доставлено",
            subscribeKey: "topic1",
            publishKey:   "topic2",
            message:      "Это сообщение не должно быть получено",
            expectMsg:    false,
        },
        {
            name:         "Пустой ключ подписки и публикации",
            subscribeKey: "",
            publishKey:   "",
            message:      "Сообщение для пустого ключа",
            expectMsg:    true,
        },
        {
            name:         "Длинное сообщение",
            subscribeKey: "long-msg",
            publishKey:   "long-msg",
            message:      "Очень длинное сообщение с повторяющимся текстом. " + 
                          "Тест тест тест тест тест тест тест тест.",
            expectMsg:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            pubSub := NewSubPub()
            
            // Используем канал и WaitGroup для синхронизации
            msgChan := make(chan interface{}, 1)
            var wg sync.WaitGroup
            wg.Add(1)
            
            // Создаем обработчик сообщений
            handler := func(msg interface{}) {
                msgChan <- msg
                wg.Done()
            }
            
            // Подписываемся на топик
            _, err := pubSub.Subscribe(tt.subscribeKey, handler)
            require.NoError(t, err)
            
            // Публикуем сообщение
            err = pubSub.Publish(tt.publishKey, tt.message)
            require.NoError(t, err)
            
            // Проверяем получение сообщения
            if tt.expectMsg {
                // Если ожидаем сообщение, ждем его получения
                select {
                case receivedMsg := <-msgChan:
                    assert.Equal(t, tt.message, receivedMsg)
                    wg.Wait() // Ждем завершения обработчика
                case <-time.After(100 * time.Millisecond):
                    t.Fatal("Таймаут: сообщение не получено")
                }
            } else {
                // Если не ожидаем сообщение, проверяем что оно не пришло
                select {
                case <-msgChan:
                    t.Fatal("Получено неожиданное сообщение")
                case <-time.After(100 * time.Millisecond):
                    // Всё хорошо, сообщение не пришло как и ожидалось
                }
                wg.Done() // Уменьшаем счетчик, т.к. обработчик не будет вызван
            }
        })
    }
}

// TestUnsubscribe проверяет корректность отписки от топика
func TestUnsubscribe(t *testing.T) {
    tests := []struct {
        name string
        key  string
    }{
        {
            name: "Базовая отписка",
            key:  "test-topic",
        },
        {
            name: "Отписка с пустым ключом",
            key:  "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            pubSub := NewSubPub()
            
            // Создаем обработчик сообщений
            handler := func(msg interface{}) {}
            
            // Подписываемся на топик
            subscription, err := pubSub.Subscribe(tt.key, handler)
            require.NoError(t, err)
            
            // Проверяем что подписка существует
            pubSub.mu.Lock()
            _, ok := pubSub.subscribers[tt.key]
            pubSub.mu.Unlock()
            assert.True(t, ok, "Подписка должна быть создана")
            
            // Отписываемся
            subscription.Unsubscribe()
            
            // Проверяем что подписчик удален
            pubSub.mu.Lock()
            subscribers, ok := pubSub.subscribers[tt.key]
            pubSub.mu.Unlock()
            if ok {
                // Если ключ все еще существует, проверяем что подписчиков нет
                assert.Equal(t, 0, len(subscribers), "Подписчики должны быть удалены")
            } else {
                // Ключ может быть полностью удален если подписчиков больше нет
                assert.True(t, true, "Ключ и подписчики удалены")
            }
        })
    }
}

// TestClose проверяет корректность закрытия pubsub
func TestClose(t *testing.T) {
    tests := []struct {
        name           string
        publishBefore  bool
        subscribeBefore bool
        timeout        time.Duration
        wantErr        bool
    }{
        {
            name:           "Базовое закрытие",
            publishBefore:  false,
            subscribeBefore: false,
            timeout:        100 * time.Millisecond,
            wantErr:        false,
        },
        {
            name:           "Закрытие с активными подписками",
            publishBefore:  false,
            subscribeBefore: true,
            timeout:        100 * time.Millisecond,
            wantErr:        false,
        },
        {
            name:           "Закрытие с таймаутом",
            publishBefore:  true,
            subscribeBefore: true,
            timeout:        1 * time.Millisecond, // очень короткий таймаут
            wantErr:        false, // таймаут может и не сработать, т.к. обработка может быть быстрой
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            pubSub := NewSubPub()
            
            if tt.subscribeBefore {
                // Создаем обработчик сообщений с задержкой для тестирования таймаута
                handler := func(msg interface{}) {
                    if tt.publishBefore {
                        time.Sleep(10 * time.Millisecond)
                    }
                }
                
                // Подписываемся на топик
                _, err := pubSub.Subscribe("test-topic", handler)
                require.NoError(t, err)
                
                if tt.publishBefore {
                    // Публикуем сообщение для проверки работы Close с активными обработчиками
                    err = pubSub.Publish("test-topic", "test message")
                    require.NoError(t, err)
                }
            }
            
            // Создаем контекст с таймаутом
            ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
            defer cancel()
            
            // Закрываем pubsub
            err := pubSub.Close(ctx)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                // Возможно ошибка, если контекст истек, но это не всегда произойдет
                if err != nil {
                    assert.ErrorIs(t, err, context.DeadlineExceeded)
                } else {
                    assert.NoError(t, err)
                }
            }
            
            // Проверяем что pubsub помечен как закрытый
            assert.True(t, pubSub.closed, "PubSub должен быть помечен как закрытый")
            
            // После закрытия подписка и публикация должны возвращать ошибки
            _, err = pubSub.Subscribe("topic", func(msg interface{}) {})
            assert.Error(t, err, "Подписка на закрытый pubsub должна вернуть ошибку")
            
            err = pubSub.Publish("topic", "message")
            assert.Error(t, err, "Публикация в закрытый pubsub должна вернуть ошибку")
        })
    }
}