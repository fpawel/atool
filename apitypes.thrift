// время UNIX в миллисекундах - количество миллисекунд, прошедших с полуночи (00:00:00 UTC) 1 января 1970 года
typedef i64 TimeUnixMillis

// Product параметры продукта
struct Product {
  1: i64 productID
  2: i64 partyID
  3: i64 serial
  4: bool active
  5: TimeUnixMillis partyCreatedAt
  6: string comport
  7: i8 addr
  8: string device
}

struct Party {
    1: i64 partyID
    2: TimeUnixMillis createdAt
    3: string name
    4: list<Product> products
}

struct PartyInfo {
    1: i64 partyID
    2: string name
    3: TimeUnixMillis createdAt
}

struct PartyParamValue {
    1: string key
    2: string name
    3: string value
}

struct ProductParamSeries {
    1: i64 productID
    2: i16 ParamAddr
    3: string chart
    4: bool seriesActive
}

const i8 NoValidate = 0;
const i8 Valid = 1;
const i8 Invalid = 2;

struct SectionProductParamsValues {
    1: string section
    2: list<string> keys
    3: list<list<string>> values
}

struct DeleteChartPointsRequest {
    1: string chart
    2: TimeUnixMillis timeFrom
    3: TimeUnixMillis timeTo
    4: double valueFrom
    5: double valueTo
}

struct AppConfig {
    1: GasDeviceConfig gas
    2: TemperatureDeviceConfig temperature
}

struct GasDeviceConfig {
    1:i8 deviceType
    2:string comport
}

struct TemperatureDeviceConfig {
    1:i8 deviceType
    2:string comport
}

struct Coefficient {
    1:i32 n
    2:bool active
}

struct ProductCoefficientValue {
    1:i64 productID
    2:i32 coefficient
    3:double value
}

struct DeviceParam {
    1:i32 ParamAddr
    2:string name
}