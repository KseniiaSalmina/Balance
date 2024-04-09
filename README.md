# API баланса пользователей
Сервис для управления балансом пользователей. Позволяет пополнять счёт, снимать деньги со счёта и осуществлять перевод от пользователя к пользователю. Хранит историю операций.

## API

API работает с форматом JSON:

    GET /wallets/{id}/balance - возвращает баланс пользователя по id.
    GET /wallets{id}/history - возвращает историю операций по id. Может принимать параметры для настройки лимита записей и сортировки (по дате или сумме, по убыванию или возрастанию). По умолчанию установена сортировка по убыванию даты и лимит в 100 записей. 
    PATCH /wallets/{id}/transaction - изменяет баланс пользователя. Поддерживает операции пополнения, снятия и перевода между пользователями.
<br>
Формат хранимых операций:

    Date        int64             //Unix timestamp
    Operation   string
    Amount      decimal.Decimal 
    Description string 

Формат запроса на изменение баланса пользователя:

    IsTransfer  bool             //reports whether transaction is a transfer or not, default false
    To          int              //required for a transfer
    Amount      decimal.Decimal  //for a transfer must be a positive number, for a not transfer transaction reports whether the operation is a replenishment (positive amount) or withdrawal (negative)
    Description string           //required for a not transfer transactions


### Создание нового счёта
При попытке пополнения или осуществления перевода на несуществующий счёт, будет создан новый счёт с указаным id.

## Переменные окружения
 Умеет считывать переменные из файла .env в директории исполняемого файла (в корне проекта).

Переменные сервера:

    SERVER_LISTEN=:8088
    SERVER_READ_TIMEOUT=5s
    SERVER_WRITE_TIMEOUT=5s
    SERVER_IDLE_TIMEOUT=30s

Переменные для подключения к Postgres:

    PG_USER=
    PG_PASSWORD=
    PG_HOST=localhost
    PG_PORT=5432
    PG_DATABASE=

В примерах указаны дефолтные значения. Если программа не сможет считать пользовательские env, то возьмет их.

