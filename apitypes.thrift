// время UNIX в миллисекундах - количество миллисекунд, прошедших с полуночи (00:00:00 UTC) 1 января 1970 года
typedef i64 TimeUnixMillis

// Product параметры продукта
struct Product {
  1: i64 productID
  2: i64 partyID
  3: TimeUnixMillis createdAt
  4: TimeUnixMillis partyCreatedAt
  5: string comport
  6: i8 addr
  7: bool checked
  8: string device
}

struct Param {
    1:i16 TheVar
    2:string Format
}

struct Party {
    1: i64 partyID
    2: TimeUnixMillis createdAt
    3: list<Product> products
    4: list<Param> params
}

struct YearMonth{
    1: i32 year
    2: i32 month
}