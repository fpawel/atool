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
    4: list<i16> vars
    5: list<ProductVar> productVars
    6: list<ProductVarTask> productVarTasks
}

struct PartyInfo {
    1: i64 partyID
    2: TimeUnixMillis createdAt
}

struct ProductVar {
    1: i64 productID
    2: i16 theVar
    3: string chart
    4: bool active
    5: string deviceVarName
}

struct ProductVarTask {
    1: i64 productID
    2: bool active
    3: string time
    4: string comport
    5: i8 addr

    6: string device
    7: string deviceTimeoutGetResponse
    8: string deviceTimeoutEndResponse
    9: string devicePause
    10: i32 deviceMaxAttemptsRead
    11: i32 deviceBaud

    12: i16 deviceVarName
    13: i64 deviceVarID
    14: i16 deviceVar
    15: i16 deviceVarSizeRead
    16: bool deviceVarMultiplyRead

    17: string request
    18: string response
    19: string duration
    20: bool ok
}
