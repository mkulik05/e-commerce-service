import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useCookies } from 'react-cookie';

export default function Item() {
  const [searchParams] = useSearchParams();
  const [item, setItem] = useState(null);
  const itemId = searchParams.get("id");
  const [cookies, setCookie] = useCookies(['shoppery']);

  useEffect(() => {
    const fetchItem = async () => {
      if (itemId) {
        const response = await fetch(`/items/item?id=${itemId}`);
        const data = await response.json();
        setItem(data);
        console.log(data)
      }
    };

    fetchItem();
  }, [itemId]);

  const addToShoppery = () => {
    const shoppery = { ...cookies.shoppery } || {};
    shoppery[itemId] = (shoppery[itemId] || 0) + 1;
    setCookie('shoppery', shoppery, { path: '/' });
  };

  const increaseAmount = () => {
    const shoppery = { ...cookies.shoppery } || {};
    if (shoppery[itemId]) {
      shoppery[itemId] += 1;
      setCookie('shoppery', shoppery, { path: '/' });
    }
  };

  const decreaseAmount = () => {
    const shoppery = { ...cookies.shoppery } || {};
    if (shoppery[itemId] && shoppery[itemId] > 0) {
      shoppery[itemId] -= 1;
      if (shoppery[itemId] === 0) delete shoppery[itemId];
      setCookie('shoppery', shoppery, { path: '/' });
    }
  };

  const deleteFromShoppery = () => {
    const shoppery = { ...cookies.shoppery } || {};
    delete shoppery[itemId];
    setCookie('shoppery', shoppery, { path: '/' });
  };
  let currentAmount = 0
  if (cookies.shoppery != undefined) {
    currentAmount = cookies.shoppery[itemId] || 0; 
  } 

  if (!item) {
    return <p></p>;
  }

  return (
    <div id="item-detail">
      <h1>{item.item_name}</h1>
      <p><strong>Price:</strong> ${item.item_price}</p>
      <p><strong>Amount Available:</strong> {item.item_amount}</p>
      <p><strong>Description:</strong> {item.item_description || "No description available."}</p>
      <p><strong>Times Bought:</strong> {item.item_bought}</p>
      <p><strong>Current Amount in Shoppery:</strong> {currentAmount}</p> 

      <div>
        <button className="shoppery-button" onClick={addToShoppery}>Add to Shoppery</button>
        <button className="shoppery-button" onClick={increaseAmount}>Increase Amount</button>
        <button className="shoppery-button" onClick={decreaseAmount}>Decrease Amount</button>
        <button className="shoppery-button" onClick={deleteFromShoppery}>Delete from Shoppery</button>
      </div>
    </div>
  );
}