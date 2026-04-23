import React from 'react';
import { createRoot } from 'react-dom/client';
import EuropeMap from '../assets/europeMap.svg';

function App() {
  return (
    <div>
      <h1>SVG Test</h1>
      <EuropeMap />
    </div>
  );
}

const root = createRoot(document.getElementById('root'));
root.render(<App />);