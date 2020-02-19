require 'mil82/data'
require 'mil82/calc_error'

local d = Device('МИЛ-82')
mil82_init_data(d)
d.Calc = mil82_calc