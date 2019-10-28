
include "apitypes.thrift"

service ProductsService {
   oneway void createNewParty()
   apitypes.Party getParty(1:i64 partyID)
   void addNewProduct()
   void deleteProduct(1:i64 productID)
   void setProduct(1:apitypes.Product product)

   list<apitypes.YearMonth> listYearMonths()
}