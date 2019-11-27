// время UNIX в миллисекундах - количество миллисекунд, прошедших с полуночи (00:00:00 UTC) 1 января 1970 года
typedef i64 TimeUnixMillis

// Product параметры продукта
struct Product {
  1: i64 productID
  2: i64 partyID
  3: bool active
  4: TimeUnixMillis partyCreatedAt
  5: string comport
  6: i8 addr
  7: string device
}

struct Party {
    1: i64 partyID
    2: TimeUnixMillis createdAt
    3: list<Product> products
    4: list<i16> paramAddresses
    5: list<ProductParam> productParams
}

struct PartyInfo {
    1: i64 partyID
    2: TimeUnixMillis createdAt
}

struct ProductParam {
    1: i64 productID
    2: i16 ParamAddr
    3: string chart
    4: bool seriesActive
}
