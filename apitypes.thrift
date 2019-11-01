// время UNIX в миллисекундах - количество миллисекунд, прошедших с полуночи (00:00:00 UTC) 1 января 1970 года
typedef i64 TimeUnixMillis

// Product параметры продукта
struct Product {
  1: i64 productID
  2: i64 partyID
  3: TimeUnixMillis createdAt
  4: TimeUnixMillis partyCreatedAt
  5: i32 port
  6: i8 addr
  7: i32 serial
  8: bool checked
  9: string device
}

struct Party {
    1: i64 partyID
    2: TimeUnixMillis createdAt
    3: list<Product> products
}

struct YearMonth{
    1: i32 year
    2: i32 month
}