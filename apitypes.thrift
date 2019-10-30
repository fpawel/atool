// время UNIX в миллисекундах - количество миллисекунд, прошедших с полуночи (00:00:00 UTC) 1 января 1970 года
typedef i64 TimeUnixMillis

// Product параметры продукта
struct Product {
  1: i64 productID
  2: i64 partyID
  3: i32 port
  4: i8 addr
  5: i32 serial
  6: bool checked
  7: string device
}

struct Party {
    1: i64 partyID
    2: TimeUnixMillis createdAt
    3: list<Product> products
}

struct DeviceVar {
    1:string device
    2:string name
    3:i16 addr
    4:string format
    5:bool checked
}

struct Interrogate {
    1:string device
    2:i16 addr
    3:i16 count
    4:bool checked
}

struct YearMonth{
    1: i32 year
    2: i32 month
}