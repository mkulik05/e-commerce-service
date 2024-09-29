import { useEffect, useState } from "react";
import { useSearchParams, Form } from "react-router-dom";

export default function Item() {
  const [searchParams] = useSearchParams();
  const [item, setItem] = useState(null);
  const itemId = searchParams.get("id");

  useEffect(() => {
    const fetchItem = async () => {
      if (itemId) {
        const response = await fetch(`/items/item?id=${itemId}`);
        const data = await response.json();
        setItem(data);
      }
    };

    fetchItem();
  }, [itemId]);

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

      
    </div>
  );
}