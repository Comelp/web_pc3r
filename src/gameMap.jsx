import React, {Component} from 'react';
import EuropeMap from '../assets/europeMap.svg';

export default class GameMap extends Component {
    
    GameBoard() {
        const handleCountryClick = (e) => {
            const el = e.nativeEvent.target.closest('[data-country]');

            if (el) {
                console.log("Pays cliqué :", el.dataset.country);
            }
        };

    return (
        <div>
            <EuropeMap 
                onClick={handleCountryClick} 
                style={{ width: '90%', height: 'auto' }} 
            />
        </div>
    );
}

    render() {
        return (
        <div>
        <h1>SVG Test</h1>
        {this.GameBoard()}
        </div>
    );
    }
}