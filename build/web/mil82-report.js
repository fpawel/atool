const sections = [
    ['test_t_norm', 'НКУ', null],
    ['test_t_low', 'низкая температура', 20],
    ['test_t_high', 'высокая температура', 20],
    ['test2', 'возврат НКУ', null],
    ['test_t80', '80⁰C', 80],
    ['tex1', 'Перед технологическим прогоном', null],
    ['tex2', 'После технологического прогона', null],
];

const gases = [1, 3, 4];

const getConcentrationErrorLimit20 = (concentration, prodType) => {
    if (prodType.gas !== "CO2") {
        return 2.5 + 0.05 * concentration;
    }
    switch (prodType.scale) {
        case 4:
            return 0.2 + 0.05 * concentration;
        case 10:
            return 0.5;
        case 20:
            return 1;
        default:
            return NaN;
    }
};

const getProductConcentrationError = ({party, product, gas, ptKey, tNorm,}) => {

    const prodType = productTypes[party.ProductType];

    if (!prodType) {
        throw "invalid party.ProductType: " + party.ProductType
    }

    const value = product.Values[`${ptKey}_gas${gas}_var0`];
    const externalGasConcentration = party.Values[`c${gas}`];

    const var2 = product.Values[`${ptKey}_gas${gas}_var2`];
    let nominal = externalGasConcentration;

    let absErrLimit = getConcentrationErrorLimit20(externalGasConcentration, prodType);

    if (tNorm !== null) {
        switch (tNorm) {
            case 20:
                nominal = product.Values[`test_t_norm_gas${gas}_var0`];
                break;
            case 80:
                nominal = product.Values[`test_t80_gas${gas}_var0`];
                break;
            default:
                throw `tNorm must not be 20 or 80: ${tNorm}`;
        }

        if (prodType.gas === "CO2") {
            absErrLimit = 0.5 * Math.abs(absErrLimit * (var2 - tNorm)) / 10;
        } else {
            if (gas === 1) {
                absErrLimit = 5;
            } else {
                absErrLimit = Math.abs(nominal * 0.15);
            }
        }
    }

    const absErr = value - nominal;

    const relErrPercent =
        100 * absErr / absErrLimit;
    const ok = Math.abs(absErr) < Math.abs(absErrLimit);
    return {
        value,
        absErr,
        absErrLimit,
        relErrPercent,
        ok,
        var2,
        externalGasConcentration,
        nominal,
    }
};


const calculateProductsValues = (party) => {
    party.Products.forEach((product) => {
        if (product.Calc) {
            throw `product.Calc must not be set: ${product}`
        }
        product.Calc = {};
        sections.forEach(([ptKey, _, tNorm], pt_index) => {
            product.Calc[ptKey] = {};
            gases.forEach((gas) =>
                product.Calc[ptKey][gas] =
                    getProductConcentrationError({party, product, gas, ptKey, tNorm,})
            );
        });
    })
};

const roundOf = (n, p) => {
    const n1 = n * Math.pow(10, p + 1);
    const n2 = Math.floor(n1 / 10);
    if (n1 >= (n2 * 10 + 5)) {
        return (n2 + 1) / Math.pow(10, p);
    }
    return n2 / Math.pow(10, p);
};


const ProductValueDetail = (props) => {
    const {product, section, gas, ptKey} = props;
    const {
        value,
        nominal,
        absErr,
        absErrLimit,
        ok,
        var2,
        externalGasConcentration
    } = product.Calc[ptKey][gas];

    const okColor = ok ? "blue" : "red";

    const round3 = (x) => roundOf(x, 3);

    return <table>
        <thead>
        <tr>
            <td colSpan={2}>{section}</td>
        </tr>
        </thead>
        <tbody>
        <tr>
            <td >Серийный №:</td>
            <td >{product.Serial}</td>
        </tr>

        <tr>
            <td>ПГС{gas}:</td>
            <td>{externalGasConcentration}</td>
        </tr>
        {externalGasConcentration !== nominal ?
            <tr>
                <td>Номинал:</td>
                <td>{nominal}</td>
            </tr> : null
        }

        <tr>
            <td>Температура:</td>
            <td>{roundOf(var2, 1)} ⁰C</td>
        </tr>
        <tr>
            <td>Макс.погр.:</td>
            <td>{round3(absErrLimit)}</td>
        </tr>
        <tr style={{color: okColor}}>
            <td>Концентрация:</td>
            <td>{round3(value)}</td>
        </tr>
        <tr style={{color: okColor}}>
            <td>Погрешность:</td>
            <td>{round3(absErr)}</td>
        </tr>
        </tbody>
    </table>;
};

const Overlay = ({hide, children}) => {
    return [
        <div className="overlay" onClick={() => hide()} key={1}/>,
        <div className="overlay-text" onClick={() => hide()} key={2}>
            {children}
        </div>
    ];
};

const Report = ({setOverlay, setProductValueDetail, party, prodType}) => {

    const products = party.Products;
    return sections.map(([ptKey, section, tNorm], pt_index) => {
        let hasValue;
        gases.forEach((gas) =>
            products.forEach((product) => {
                const {relErrPercent} = product.Calc[ptKey][gas];
                if (relErrPercent === relErrPercent) {
                    hasValue = true;
                }
            })
        );
        if (!hasValue) {
            return null;
        }
        return <section key={ptKey}>
            <h2>{section}</h2>
            <table>
                <thead>
                <tr>
                    <th>№</th>
                    {
                        products.map((product) =>
                            <th key={product.ProductID}> {product.Serial} </th>)
                    }
                </tr>
                </thead>
                <tbody>
                {
                    [1, 3, 4].map((gas) => <tr key={gas}>
                        <th>ПГС{gas}</th>
                        {
                            products.map((product) => {
                                const x = {party, product, prodType, gas, ptKey, tNorm, section};
                                const {relErrPercent, ok} = product.Calc[ptKey][gas];
                                if (relErrPercent !== relErrPercent) {
                                    return <td key={product.ProductID}/>;
                                }
                                return <td
                                    key={product.ProductID}
                                    className={ok ? "ok" : "err"}
                                    onClick={
                                        () => {
                                            setProductValueDetail({...x});
                                            setOverlay(true);
                                        }
                                    }
                                >
                                    {roundOf(relErrPercent, 1)}
                                </td>;
                            })
                        }
                    </tr>)
                }
                </tbody>
            </table>
        </section>;
    });
};

const App = () => {
    const [party, setParty] = React.useState(null);
    const [overlay, setOverlay] = React.useState(false);
    const [productValueDetail, setProductValueDetail] = React.useState(null);
    const [failed, setFailed] = React.useState(null);

    if (failed) {
        return <div>
            <h1 style={{color:"red"}}>{failed}</h1>,
            <a href={"/"}>Вернуться на главную страницу</a>
        </div>;
    }

    if (party === null) {
        (async () => {
            const response = await fetch('/party');
            let party = await response.json();
            window.party = party;
            document.title = `Партия ${party.PartyID}`;
            if (!party.Products) {
                setFailed("Нет данных!")
                return
            }
            calculateProductsValues(party);
            setParty(party);
        })();
        return <h1>получение данных...</h1>;
    }

    const props = {party, setOverlay, setProductValueDetail, key:1};
    return [
        <h1 key={0} style={{borderBottom: "1px solid #BCBCBC", paddingBottom: "10px"}}>Расчёт погрешности МИЛ-82</h1>,
        <Report {...props } />,
        <h4 style={{textAlign: "left"}} key={2}>
            ПГС1: <strong style={{marginRight: "30px"}}>{party.Values.c1}</strong>
            ПГС3: <strong style={{marginRight: "30px"}}>{party.Values.c3}</strong>
            ПГС4: <strong>{party.Values.c4}</strong>
        </h4>,
        overlay ?
            <Overlay hide={() => setOverlay(false)} key={3}>
                <ProductValueDetail {...productValueDetail}/>
            </Overlay>
            : null,
    ];
};


ReactDOM.render(
    <App/>,
    document.getElementById('root')
);

