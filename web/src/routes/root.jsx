import React, { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate } from "react-router-dom";

export default function Shop() {
  const location = useLocation();
  const navigate = useNavigate();

  const [searchQuery, setSearchQuery] = useState(
    new URLSearchParams(location.search).get('search') || ''
  );
  const [items, setItems] = useState([]);
  const [currentPage, setCurrentPage] = useState(
    Number(new URLSearchParams(location.search).get('page')) || 0
  );
  const [totalPages, setTotalPages] = useState(0);
  const [sortCriteria, setSortCriteria] = useState(
    new URLSearchParams(location.search).get('sort') || 'name'
  );
  const [sortOrder, setSortOrder] = useState(
    new URLSearchParams(location.search).get('sort_order') || 'asc'
  );

  const fetchItems = async () => {
    const response = await fetch(`/items/list?search=${encodeURIComponent(searchQuery)}&page=${currentPage}&sort=${sortCriteria}&sort_order=${sortOrder}`);
    const data = await response.json();
    if (data) {
      setTotalPages(data.amount);
      setItems(data.items);
    }
  };

  const handleSearch = async (event) => {
    event.preventDefault();
    setCurrentPage(0);
    navigate(`?search=${encodeURIComponent(searchQuery)}&page=0&sort=${sortCriteria}&sort_order=${sortOrder}`);
    await fetchItems();
  };

  const handlePageChange = async (page) => {
    setCurrentPage(page);
    navigate(`?search=${encodeURIComponent(searchQuery)}&page=${page}&sort=${sortCriteria}&sort_order=${sortOrder}`);
    await fetchItems();
  };

  useEffect(() => {
    fetchItems();
  }, [currentPage, sortCriteria, sortOrder]);

  return (
    <div id="shop" className="wrapper">
      <header>
        <h1>Shop</h1>
        <Link to={`/shoppery`}>Shoppery</Link>
      </header>
      <form id="search-form" role="search" onSubmit={handleSearch}>
        <input
          id="q"
          aria-label="Search shop items"
          placeholder="Search"
          type="search"
          name="q"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
        <button type="submit">Search</button>
      </form>

      <div className="sorting-container">
        <label htmlFor="sort">Sort By:</label>
        <select
          id="sort"
          value={sortCriteria}
          onChange={(e) => {
            setSortCriteria(e.target.value);
            navigate(`?search=${encodeURIComponent(searchQuery)}&page=${currentPage}&sort=${e.target.value}&sort_order=${sortOrder}`);
          }}
        >
          <option value="name">Name</option>
          <option value="price">Price</option>
          <option value="times_bought">Times Bought</option>
        </select>
        
        <label htmlFor="sortOrder">Order:</label>
        <select
          id="sortOrder"
          value={sortOrder}
          onChange={(e) => {
            setSortOrder(e.target.value);
            navigate(`?search=${encodeURIComponent(searchQuery)}&page=${currentPage}&sort=${sortCriteria}&sort_order=${e.target.value}`);
          }}
        >
          <option value="asc">Ascending</option>
          <option value="desc">Descending</option>
        </select>
      </div>

      <main>
        <section id="item-list">
          <ul>
            {items.map(item => (
              <li key={item.id}>
                <Link to={`/item?id=${item.id}`} className="item-link">
                  <span className="item-name">{item.name}</span>
                  <span className="item-price">${item.price}</span>
                </Link>
              </li>
            ))}
          </ul>
        </section>
        <footer className="pagination">
          {Array.from({ length: totalPages }, (_, index) => (
            <button key={index} onClick={() => handlePageChange(index)} disabled={currentPage === index}>
              {index + 1}
            </button>
          ))}
        </footer>
      </main>
    </div>
  );
}