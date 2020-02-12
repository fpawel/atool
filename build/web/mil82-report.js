const getProductConcentrationError = ({party, product, prodType, gas, ptKey, tNorm}) => {
    const value = product.Values[`${ptKey}_gas${gas}_var0`];
    const nominal = party.Values[`c${gas}`];
    const absErr = value - nominal;

    let absErrLimit = concentrationErrorLimit20(value, prodType);
    let var2 = null;
    let dTempErr = null;
    if (tNorm !== null) {
        var2 = product.Values[`${ptKey}_gas${gas}_var2`];
        const x = 0.5 * Math.abs(absErrLimit * (var2 - tNorm)) / 10;
        dTempErr = x - absErrLimit;
        absErrLimit = x;

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
        dTempErr,
    }
};

const ProductValueDetail = (props) => {
    const {
        value,
        nominal,
        absErr,
        absErrLimit,
        ok,
        var2,
        dTempErr,
    } = getProductConcentrationError(props);

    const {product, section, gas} = props;

    const okColor = ok ? "blue" : "red";
    return [
        <h4 >
            МИЛ-82: {product.Serial}
        </h4>,
        <h4 style={{
            borderBottom: "1px solid black",
            paddingBottom: "20px",
        }}>
            {section}
        </h4>,
        <table>
            <tr>
                <td>ПГС{gas}:</td>
                <td>{nominal}</td>
            </tr>
            <tr>
                <td>Предел абс. погрешности:</td>
                <td>{absErrLimit.toFixed(3)}</td>
            </tr>
            <tr style={{color:okColor}}>
                <td>Концентрация:</td>
                <td>{value.toFixed(3)}</td>
            </tr>
            <tr style={{color:okColor}}>
                <td>Абс. погрешность:</td>
                <td>{absErr.toFixed(3)}</td>
            </tr>
            {
                var2 ? <tr>
                    <td>var2:</td>
                    <td>{var2.toFixed(3)}</td>
                </tr> : null

            }
            {
                dTempErr ? <tr>
                    <td>dTempErr:</td>
                    <td>{dTempErr.toFixed(3)}</td>
                </tr> : null

            }
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

const Report = () => {
    const [party, setParty] = React.useState(null);
    const [overlay, setOverlay] = React.useState(false);
    const [productValueDetail, setProductValueDetail] = React.useState(null);


    if (party === null) {
        (async () => {
            const response = await fetch('/party');
            const x = await response.json();
            document.title = `Партия ${x.PartyID}`;
            setParty(x);
        })();
        return <h1>получение данных</h1>;
    }

    const products = party.Products;
    const prodType = productTypes[party.ProductType];

    const tabs = [
        ['test_t_norm', 'нормальная температура (НКУ)', null],
        ['test_t_low', 'низкая температура', 20],
        ['test_t_high', 'высокая температура', 20],
        ['test2', 'возврат НКУ', null]].map(([ptKey, section, tNorm], pt_index) =>
        <table className="report-table">
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
                        <th key={product.ProductID}> {product.Serial} </th> )
                }
            </tr>
            </thead>
            <tbody>
            {
                [1, 3, 4].map( (gas) => <tr key={gas}>
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
                                {relErrPercent.toFixed(1)}
                            </td>;
                        })
                    }
                </tr>)
            }
            </tbody>
        </table>
    );

    const main = <div className={"centered"}>
        {tabs}
    </div>;

    let overlayElement = overlay ?
        <Overlay hide = {() => setOverlay(false)} >
            <ProductValueDetail {...productValueDetail}/>
        </Overlay>
        : null;

    return [overlayElement, main];
};


ReactDOM.render(
    <Report/>,
    document.getElementById('root')
);

