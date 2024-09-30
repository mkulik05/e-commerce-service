import { useEffect, useState } from "react";
import { useCookies } from 'react-cookie';

export default function Shoppery() {
  const [cookies, setCookie] = useCookies(['shoppery']);
  const [shopperyItems, setShopperyItems] = useState({});

  useEffect(() => {
    const shoppery = cookies.shoppery || {};
    setShopperyItems(shoppery);
  }, [cookies]);

  const updateAmount = (itemId, amount) => {
    const shoppery = { ...shopperyItems };
    if (amount > 0) {
      shoppery[itemId] = amount;
    } else {
      delete shoppery[itemId];
    }
    setCookie('shoppery', shoppery, { path: '/' });
    setShopperyItems(shoppery);
  };

  const deleteFromShoppery = (itemId) => {
    const shoppery = { ...shopperyItems };
    delete shoppery[itemId];
    setCookie('shoppery', shoppery, { path: '/' });
    setShopperyItems(shoppery);
  };

  return (
    <div id="shoppery">
      <h1>Your Shoppery</h1>
      {Object.entries(shopperyItems).length === 0 ? (
        <p>Your shoppery is empty.</p>
      ) : (
        Object.entries(shopperyItems).map(([itemId, amount]) => (
          <div key={itemId} className="shoppery-item">
            <p>Item ID: {itemId}</p>
            <input
              type="number"
              value={amount}
              min="0"
              onChange={(e) => updateAmount(itemId, Number(e.target.value))}
            />
            <button onClick={() => deleteFromShoppery(itemId)}>Delete</button>
          </div>
        ))
      )}
    </div>
  );
}