const MapProducts = ({products, component}) => {
    const Component = component;
    return products.map((p) => {
        return <Component key={p.ProductID} product={p}/>;
    })
};

const ProductsSerialsRow = ({products}) => {
    return <MapProducts
        products={products}
        component={
            ({product}) => <th> {product.Serial} </th>
        }/>;
};


const ProductConcentrationError = ({value, nominal, prodType, product, setOverlay}) => {
    const absErr = value - nominal;
    const absErrLimit = concentrationErrorLimit20(value, prodType);
    const relErrPercent =
        100 * Math.abs(absErr) / absErrLimit;
    if (relErrPercent !== relErrPercent) {
        return <td/>;
    }
    const ok = Math.abs(absErr) < Math.abs(absErrLimit);
    const className = ok ? "ok" : "err";

    return <td className={className}
               onClick={() =>
                   setOverlay({
                       serial:product.Serial,
                       value:value,
                       nominal:nominal,
                       absErr:absErr,
                       absErrLimit:absErrLimit,
                       ok:ok,
                   })
               }
    >
        {relErrPercent.toFixed(1)}
    </td>
};


const Report = () => {

    const [party, setParty] = React.useState(null);
    const [overlayState, setOverlay] = React.useState(null);

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

    const xs = [
        ['test_t_norm', 'нормальная температура (НКУ)', null],
        ['test_t_low', 'низкая температура', 20],
        ['test_t_high', 'высокая температура', 20],
        ['test2', 'возврат НКУ', null]].map(([ptKey, section, tNorm]) => {
        return <div className={"centered"}>
            <table className="report-table">
                <caption style={{textAlign:"left", fontSize:"16px", margin:"5px"}}>{section}</caption>
                <thead>
                <tr>
                    <th>Газ</th>
                    <ProductsSerialsRow products={products}/>
                </tr>
                </thead>
                <tbody>
                {
                    [1, 3, 4].map((gas) => (
                        <tr key={gas}>
                            <td>ПГС{gas}</td>
                            <MapProducts
                                products={products}
                                component={
                                    ({product}) =>
                                         <ProductConcentrationError
                                            product={product}
                                            value={product.Values[`${ptKey}_gas${gas}_var0`]}
                                            prodType={prodType}
                                            nominal={party.Values[`c${gas}`]}
                                            setOverlay={setOverlay}
                                         />
                                }/>

                        </tr>
                    ))
                }
                </tbody>
            </table>
        </div>;

    });

    return <div>
        <OverlayProductConcentrationError overlayState={overlayState} hide={() => setOverlay(null)} />
        ПГС1=<strong style={{marginRight: "10px"}}>{party.Values.c1}</strong>
        ПГС3=<strong style={{marginRight: "10px"}}>{party.Values.c3}</strong>
        ПГС4=<strong>{party.Values.c4}</strong>
        {xs}
    </div>;
};

const OverlayProductConcentrationError = ({overlayState, hide}) => {

    if (!overlayState) {
        return null;
    }
    let {serial, value, nominal, absErr, absErrLimit, ok} = overlayState;

    const className = ok ? "ok" : "err";
    return [
        <div className="overlay" onClick={() => hide()}/>,
        <div className="overlay-text" onClick={() => hide()} >
            <h3 style={{
                borderBottom: "1px solid black",
                paddingBottom: "20px",
            }}>
                МИЛ-82: {serial}
            </h3>
            <table>
                <tr>
                    <td>ПГС:</td>
                    <td>{nominal.toFixed(3)}</td>
                </tr>
                <tr>
                    <td>Предел абс. погрешности:</td>
                    <td>{absErrLimit.toFixed(3)}</td>
                </tr>
                <tr>
                    <td>Концентрация:</td>
                    <td className={className} >{value.toFixed(3)}</td>
                </tr>
                <tr>
                    <td>Абс. погрешность:</td>
                    <td className={className}>{absErr.toFixed(3)}</td>
                </tr>


            </table>
        </div>
    ]
};


ReactDOM.render(
    <Report/>,
    document.getElementById('root')
);

