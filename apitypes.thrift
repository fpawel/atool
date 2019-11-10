// время UNIX в миллисекундах - количество миллисекунд, прошедших с полуночи (00:00:00 UTC) 1 января 1970 года
typedef i64 TimeUnixMillis

// Product параметры продукта
struct Product {
  1: i64 productID
  2: i64 partyID
  3: TimeUnixMillis partyCreatedAt
  4: string comport
  5: i8 addr
  6: bool checked
  7: string device
}

struct Party {
    1: i64 partyID
    2: TimeUnixMillis createdAt
    3: list<Product> products
    4: list<i16> params
    5: list<string> charts
    6: list<ParamVarSeries> series
}

struct PartyInfo {
    1: i64 partyID
    2: TimeUnixMillis createdAt
}

struct ParamVarSeries {
    1: i64 productID
    2: i16 theVar
    3: string chart
    4: string color
}