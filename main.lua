function sleep(n)  -- seconds
    local t0 = os.clock()
    while os.clock() - t0 <= n do end
end

print("lua: пауза1")
delay_sec(20, "пауза1")
print("lua: пауза2")
delay_sec(25, "пауза2")
print("lua: продувка")
delay_sec(25, "пауза3")