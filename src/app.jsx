import React from 'react';
import { createRoot } from 'react-dom/client';
import GameMap from './gameMap';
import Timer from './timer';
import Login from './login';
import Rules from "./rules";
import UserInterface from './userinterface';
import { AuthProvider } from './authContext';

function App() {
  return (
  <AuthProvider>
    <div>
      <div>
        <h1>Bienvenue dans le jeu de conquête de l'Europe !</h1>
      </div>
      <Rules />
      <Login />
      <UserInterface />
      <Timer />
      <GameMap />
    </div>
  </AuthProvider>
  );
}

const root = createRoot(document.getElementById('root'));
root.render(<App />);