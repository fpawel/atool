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


const ProductConcentrationError = ({value, nominal, prodType}) => {
    const absErr = value - nominal;
    const absErrLimit = concentrationErrorLimit20(value, prodType);
    const relErrPercent =
        100 * Math.abs(absErr) / absErrLimit;
    if (relErrPercent !== relErrPercent) {
        return <td/>;
    }
    const className = Math.abs(absErr) < Math.abs(absErrLimit) ? "ok" : "err";
    return <td className={className}
               onClick={() => {
                   let s = [
                       ["Концентрация", value],
                       ["ПГС", nominal],
                       ["Абсолютная погрешность", absErr],
                       ["Предел абсолютной погрешности", absErrLimit]].map(([s, v]) => `${s}: ${v.toFixed(3)}`).join("\n");
                   alert(s)
               }}
    >
        {relErrPercent.toFixed(1)}
    </td>
};

const Report = () => {

    const [party, setParty] = React.useState(null);

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
        return <div>
            <h3>{section}</h3>
            <table className="report-table">
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
                                    ({product}) => {
                                        const value = product.Values[`${ptKey}_gas${gas}_var0`];
                                        const nominal = party.Values[`c${gas}`];
                                        return <ProductConcentrationError
                                            value={value}
                                            prodType={prodType}
                                            nominal={nominal}/>
                                    }
                                }/>

                        </tr>
                    ))
                }
                </tbody>
            </table>
        </div>;

    });

    return <div>
        ПГС1=<strong style={{marginRight: "10px"}}>{party.Values.c1}</strong>
        ПГС3=<strong style={{marginRight: "10px"}}>{party.Values.c3}</strong>
        ПГС4=<strong>{party.Values.c4}</strong>
        {xs}
    </div>;
};

// const App = () => (<HashRouter>
//     <ul>
//         <li><Link to="/">МИЛ-82: погрешность</Link></li>
//         <li><Link to="/a">TO A</Link></li>
//         <li><Link to="/b">TO B</Link></li>
//     </ul>
//     <Route path="/" exact component={Mil82Report}/>
//     <Route path="/a" component={A}/>
//     <Route path="/b" component={B}/>
// </HashRouter>);


ReactDOM.render(
    <Report/>,
    document.getElementById('root')
);

