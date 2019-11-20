// время UNIX в миллисекундах - количество миллисекунд, прошедших с полуночи (00:00:00 UTC) 1 января 1970 года
typedef i64 TimeUnixMillis

// Product параметры продукта
struct Product {
  1: i64 productID
  2: i64 partyID
  3: TimeUnixMillis partyCreatedAt
  4: string comport
  5: i8 addr
  6: string device
}

struct Party {
    1: i64 partyID
    2: TimeUnixMillis createdAt
    3: list<Product> products
    4: list<i16> params
    5: list<ProductVarSeries> series
    6: list<ProductVarTask> productVarTasks
}

struct PartyInfo {
    1: i64 partyID
    2: TimeUnixMillis createdAt
}

struct ProductVarSeries {
    1: i64 productID
    2: i16 theVar
    3: string chart
    4: bool active
}

struct ProductVarTask {
    1: bool check
    2: string time
    3: string comport
    4: i64 productID
    5: string device
    6: i8 addr
    7: i16 theVar
    8: i16 count
    9: string request
    10: string response
    11: string duration
    12: bool ok
}
