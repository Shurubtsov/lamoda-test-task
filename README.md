# Тестовое задание от компании Lamoda

## Описание

Необходимо спроектировать и реализовать API методы для работы с товарами на
одном складе.
Учесть, что вызов API может быть одновременно из разных систем и они могут
работать с одинаковыми товарами.

## Запуск

Для запуска решения требуется наличие `docker-compose` и `make`

Команда для поднятия:
```bash
make up
```

Для завершения и очистки:
```bash
make down
```

Дефолтная конфигурация:

```yaml
# Переменные окружения использующиеся в конфиге
      ADDRESS: "0.0.0.0:8082" # адресс на котором запускается http сервер
      MIGRATION_VERSION: 2 # версия миграции
      MIGRATIONS_PATH: file://./ # путь к файлам миграции
      LEVEL: -1 # уровень логгирования (описан в pkg) , значение от -1 до 4
      OUTPUT: dev # вывод логов в stderr или io.discard, значения "dev" и "prod"
      # переменные для бд
      DB_USERNAME: postgres
      DB_PASSWORD: admin
      DB_PORT: "5432"
      DB_HOST: db
      DB_DATABASE: go_test_db
```

Контейнеры:

1. Само приложение app

2. Инстанс PostgreSQL

3. Adminer для администрирования бд вручную и наглядных проверок результатов

## Ручное тестирование

- резервирование товара на складе для доставки

Запрос:
```bash
curl -X POST http://0.0.0.0:8082/product/reservation \
-H "Content-Type: application/json" \
-d '[
    {"code": "ID-SN"},
    {"code": "CZ-ZL"},
    {"code": "MM-17"},
    {"code": "MW-KR"},
    {"code": "CA-ON"},
    {"code": "AU-NSW"}
]'
```

Ответ:
```json
{

"message": "reservation successful complete",

"reserved_products": [

    {

        "code": "ID-SN",

        "name": "Bread - Ciabatta Buns",

        "id": 1,

        "size": 1,

        "count": 21

    },

    {

        "code": "CZ-ZL",

        "name": "Cookies - Oreo, 4 Pack",

        "id": 13,

        "size": 13,

        "count": 25

    },

    {

        "code": "MM-17",

        "name": "Skewers - Bamboo",

        "id": 19,

        "size": 19,

        "count": 6

    },

    {

        "code": "MW-KR",

        "name": "Squid Ink",

        "id": 44,

        "size": 44,

        "count": 28

    },

    {

        "code": "CA-ON",

        "name": "Wine - Chateauneuf Du Pape",

        "id": 40,

        "size": 40,

        "count": 8

    },

    {

        "code": "AU-NSW",

        "name": "Milk - Chocolate 250 Ml",

        "id": 11,

        "size": 11,

        "count": 8

    }

],

"status": "OK"

}
```

- освобождение резерва товаров

Запрос:
```bash
curl -X DELETE http://0.0.0.0:8082/product/exemption \
-H "Content-Type: application/json" \
-d '[
    {"code": "ID-SN"},
    {"code": "CZ-ZL"},
    {"code": "MM-17"},
    {"code": "MW-KR"},
    {"code": "CA-ON"},
    {"code": "AU-NSW"}
]'
```

Ответ:
```json
{
  "exempted_products": [
    {
      "code": "ID-SN",
      "name": "Bread - Ciabatta Buns",
      "id": 1,
      "size": 1,
      "count": 21
    },
    {
      "code": "CZ-ZL",
      "name": "Cookies - Oreo, 4 Pack",
      "id": 13,
      "size": 13,
      "count": 25
    },
    {
      "code": "MM-17",
      "name": "Skewers - Bamboo",
      "id": 19,
      "size": 19,
      "count": 6
    },
    {
      "code": "MW-KR",
      "name": "Squid Ink",
      "id": 44,
      "size": 44,
      "count": 28
    },
    {
      "code": "CA-ON",
      "name": "Wine - Chateauneuf Du Pape",
      "id": 40,
      "size": 40,
      "count": 8
    },
    {
      "code": "AU-NSW",
      "name": "Milk - Chocolate 250 Ml",
      "id": 11,
      "size": 11,
      "count": 8
    }
  ],
  "message": "exemption successful complete",
  "status": "OK"
}
```

- получение кол-ва оставшихся товаров на складе

Запрос:
```bash
curl -X GET http://0.0.0.0:8082/storage/products?id=1
```

Ответ:
```json
{

"count_all_products": 30,

"message": "successful getting all remaining products from storage",

"remaining_products": [

	{
	
		"code": "CO-MET",
		
		"name": "Roe - Lump Fish, Red",
		
		"id": 34,
		
		"size": 34,
		
		"count": 30
	
	}

],

"status": "OK"

}
```

### Комментарии

Так как условие тестового задания подразумевает недосказанность, оттого есть пробелы в моей реализации

В первом методе я сделал ставку на то, что товары уже заведены в базу данных, поэтому решил только основываться на валидации и поиске крайних случаев. Есть что улучшить, но в целом как есть.

Второй метод аналогичен первому, за исключением того что не совсем понял как лучше отдавать ответ о том что все удалено успешно.

Третий метод тоже имеет под собой пищу для размышления, но исходя из условий решил не играть в рулетку и отобразить как общее количество товаров зарезервированных за складом, так и в целом данные по каждому из товаров, находящихся на резервировании в данный момент.

Так же с моей стороны маловато указано команд для проверки в разделе `Ручное тестирование` , но я вместе с приложением решил поднять `adminer` , чтобы можно было наглядно проверять базу данных.

