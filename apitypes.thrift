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
  8: i32 place

}

struct Party {
    1: i64 partyID
    2: TimeUnixMillis createdAt
    3: string name
    4: string deviceType
    5: string productType
    6: list<Product> products

}

struct PartyInfo {
    1: i64 partyID
    2: string name
    3: string deviceType
    4: string productType
    5: TimeUnixMillis createdAt
}

struct ConfigParamValue {
    1: string key
    2: string name
    3: string type
    4: list<string> valuesList
    5: string value
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


struct PartyProductsValues {
    1:list<Product> products;
    2:list<SectionProductParamsValues> sections;
    3: list<CalcSection> calc
    4: string calcError
}

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

struct Coefficient {
    1:i32 n
    2:bool active
    3:string name
}

struct DeviceInfo {
    1:list<string> productTypes
    2:list<string> commands
    3:list<Coefficient> coefficients
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

struct CalcSection {
	1:string name
	2:list<CalcParam> params
}

struct CalcParam {
	1:string name
	2:list<CalcValue> values
}

struct CalcValue {
    1:bool validated
    2:bool	valid
    3:string  detail
    4: string value
}

struct JournalEntry {
    1:i64 entryID
    2:string storedAt
    4:i64 indent
    5:bool	ok
    6:string  text
    7:string stack
}

struct BuildInfo {
    1:string commit
    2:string commitGui
    3:string uuid
    4:string date
    5:string time
}

struct WorkLogRecord {
    1:string workName
    2:string strtedAt
    3:string completedAt
}

typedef i16 CmdModbus

struct RequestDeviceCommand {
    1:i16 cmdModbus
    2:string cmdDevice
    3:string format
    4:string argument
}