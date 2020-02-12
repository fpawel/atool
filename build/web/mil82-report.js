const getProductConcentrationError = ({party, product, gas, ptKey, tNorm, prodType}) => {
    const value = product.Values[`${ptKey}_gas${gas}_var0`];
    const nominal = party.Values[`c${gas}`];
    const absErr = value - nominal;

    let absErrLimit = concentrationErrorLimit20(nominal, prodType);
    const var2 = product.Values[`${ptKey}_gas${gas}_var2`];
    if (tNorm !== null) {
        if (prodType.gas === "CO2"){
            absErrLimit = 0.5 * Math.abs(absErrLimit * (var2 - tNorm)) / 10;
        } else {
            if (gas===1) {
                absErrLimit = 5;
            } else {
                const c20 = product.Values[`test_t_norm_gas${gas}_var0`];
                absErrLimit = Math.abs(c20 * 0.15);
            }

        }
    }
    const relErrPercent =
        100 * absErr / absErrLimit;
    const ok = Math.abs(absErr) < Math.abs(absErrLimit);
    return {
        value,
        nominal,
        absErr,
        absErrLimit,
        relErrPercent,
        ok,
        var2,
    }
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
    const {
        value,
        nominal,
        absErr,
        absErrLimit,
        ok,
        var2,
    } = getProductConcentrationError(props);

    const {product, section, gas} = props;

    const okColor = ok ? "blue" : "red";

    const round3 = (x) => roundOf(x, 3);

    const st1 = { style: {borderBottom:"1px solid #BCBCBC", paddingBottom:"5px"} };
    return [
        <table>
            <tr >
                <td {...st1} >Серийный №:</td>
                <td {...st1}>{product.Serial}</td>
            </tr>
            <tr>
                <td colSpan={2}>{section}</td>
            </tr>
            <tr>
                <td>ПГС{gas}:</td>
                <td>{nominal}</td>
            </tr>
            <tr>
                <td>Температура:</td>
                <td>{roundOf(var2, 1)}⁰C</td>
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
        </table>
    ]
};

const Overlay = ({hide, children}) => {
    return [
        <div className="overlay" onClick={() => hide()}/>,
        <div className="overlay-text" onClick={() => hide()}>
            {children}
        </div>
    ];
};

const Report = ({setOverlay, setProductValueDetail, party, prodType}) => {

    const products = party.Products;
    const tabs = [
        ['test_t_norm', 'НКУ', null],
        ['test_t_low', 'низкая температура', 20],
        ['test_t_high', 'высокая температура', 20],
        ['test2', 'возврат НКУ', null]].map(([ptKey, section, tNorm], pt_index) =>
        <table>
            <caption style={{textAlign: "left", fontSize: "16px", margin: "5px"}}>
                {section}
                {pt_index === 0 ?
                    <span style={{marginLeft: "30px"}}>
                        ПГС1: <strong style={{marginRight: "30px"}}>{party.Values.c1}</strong>
                        ПГС3: <strong style={{marginRight: "30px"}}>{party.Values.c3}</strong>
                        ПГС4: <strong>{party.Values.c4}</strong>
                    </span> : null
                }
            </caption>
            <thead>
            <tr>
                <th>Газ</th>
                {
                    products.map((product) =>
                        <th key={product.ProductID}> {product.Serial} </th>)
                }
            </tr>
            </thead>
            <tbody>
            {
                [1, 3, 4].map((gas) => <tr key={gas}>
                    <td>ПГС{gas}</td>
                    {
                        products.map((product) => {
                            const x = {party, product, prodType, gas, ptKey, tNorm, section};
                            const {relErrPercent, ok} = getProductConcentrationError(x);
                            if (relErrPercent !== relErrPercent) {
                                return <td/>;
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
    );

    return <div className={"centered report"}>
        {tabs}
    </div>;
};

const App = () => {
    const [data, setData] = React.useState(null);
    const [overlay, setOverlay] = React.useState(false);
    const [productValueDetail, setProductValueDetail] = React.useState(null);

    if (data === null) {
        (async () => {
            const response = await fetch('/party');
            let party = await response.json();
            document.title = `Партия ${party.PartyID}`;

            if (!party.ProductType) {
                throw "party.ProductType must be set!"
            }
            const prodType = productTypes[party.ProductType];

            if (!prodType) {
                throw "invalid party.ProductType: " + party.ProductType
            }
            setData({prodType, party});
        })();
        return <h1>получение данных</h1>;
    }

    let overlayElement = overlay ?
        <Overlay hide={() => setOverlay(false)}>
            <ProductValueDetail {...productValueDetail}/>
        </Overlay>
        : null;

    const props = {...data, setOverlay, setProductValueDetail,};
    return [<Report {...props} />, overlayElement];
};


ReactDOM.render(
    <App/>,
    document.getElementById('root')
);

