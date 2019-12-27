hx = 122

function sleep(n)  -- seconds
    local t0 = os.clock()
    while os.clock() - t0 <= n do end
end

works = {
	['пауза'] = function (n)
		print("пауза", n)
	end,
	['задержка'] = function (n)
        print("задержка", n)
    end,
    ['продувка газа'] = function (n)
        print("продувка газа", n)
    end
}

