require '/print_table'

print_table(go.Config)
print("product_type", go.Config.product_type)
print("c1", go.Config.c1)
print("c2", go.Config.c2)
print("c3", go.Config.c3)

for i,p in ipairs(go.Products) do -- для каждого прибора партии
    print(i, p.Device.Baud, p.Serial)
end