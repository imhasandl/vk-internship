package server

import (
    "context"
    "testing"
    "time"

    "github.com/imhasandl/vk-internship/protos"
    "github.com/imhasandl/vk-internship/subpub"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Мок для SubPub_SubscribeServer
type mockSubscribeServer struct {
    mock.Mock
    protos.SubPub_SubscribeServer
    ctx context.Context
}

func (m *mockSubscribeServer) Send(event *protos.Event) error {
    args := m.Called(event)
    return args.Error(0)
}

func (m *mockSubscribeServer) Context() context.Context {
    return m.ctx
}

// Тест для метода Publish
func TestPublish(t *testing.T) {
    // Тестовые случаи
    tests := []struct {
        name    string
        key     string
        data    string
        wantErr bool
    }{
        {
            name:    "Успешная публикация",
            key:     "test-topic",
            data:    "test message",
            wantErr: false,
        },
        {
            name:    "Пустой ключ",
            key:     "",
            data:    "test message",
            wantErr: false,
        },
        {
            name:    "Пустые данные",
            key:     "test-topic",
            data:    "",
            wantErr: false,
        },
    }

    // Выполнение тестов
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Создаем экземпляр PubSub и сервера
            pubSub := subpub.NewSubPub()
            server := NewServer("test-port", pubSub)
            
            // Вызываем метод Publish
            req := &protos.PublishRequest{
                Key:  tt.key,
                Data: tt.data,
            }
            _, err := server.Publish(context.Background(), req)
            
            // Проверяем результат
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

// Тест для метода Subscribe
func TestSubscribe(t *testing.T) {
    tests := []struct {
        name      string
        key       string
        message   string
        expectMsg bool
    }{
        {
            name:      "Базовый тест подписки",
            key:       "test-key",
            message:   "тестовое сообщение",
            expectMsg: true,
        },
        {
            name:      "Пустой ключ",
            key:       "",
            message:   "сообщение для пустого ключа",
            expectMsg: true,
        },
        {
            name:      "Специальные символы в ключе",
            key:       "test-key!@#$%^&*()",
            message:   "сообщение для ключа со спецсимволами",
            expectMsg: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Настройка контекста с возможностью отмены
            ctx, cancel := context.WithCancel(context.Background())
            defer cancel()

            // Создаем мок для стрима
            mockStream := &mockSubscribeServer{
                ctx: ctx,
            }
            
            // Настраиваем ожидание вызова Send с нашим сообщением
            mockStream.On("Send", &protos.Event{
                Data: tt.message,
            }).Return(nil)

            // Создаем экземпляр PubSub и сервера
            pubSub := subpub.NewSubPub()
            server := NewServer("test-port", pubSub)

            // Запускаем Subscribe в отдельной горутине
            errCh := make(chan error)
            go func() {
                err := server.Subscribe(&protos.SubscribeRequest{
                    Key: tt.key,
                }, mockStream)
                errCh <- err
            }()

            // Даем время на установку подписки
            time.Sleep(50 * time.Millisecond)

            // Публикуем сообщение
            err := pubSub.Publish(tt.key, tt.message)
            assert.NoError(t, err)

            // Даем время на обработку сообщения
            time.Sleep(50 * time.Millisecond)

            // Отменяем контекст, чтобы завершить Subscribe
            cancel()

            // Проверяем, что Subscribe завершился без ошибок
            err = <-errCh
            assert.ErrorIs(t, err, context.Canceled)

            // Проверяем, что mockStream.Send был вызван как ожидалось
            mockStream.AssertExpectations(t)
        })
    }
}