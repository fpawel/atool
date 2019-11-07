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
  8: list<Series> series
}

struct Series {
    1:i16 theVar
    2:i32 chartID
    3:string color
}

struct Party {
    1: i64 partyID
    2: TimeUnixMillis createdAt
    3: list<Product> products
    4: list<i16> params
    5: string note
    6: list<i32> charts
}

struct PartyInfo {
    1: i64 partyID
    2: TimeUnixMillis createdAt
    3: string note
}

struct YearMonth{
    1: i32 year
    2: i32 month
}