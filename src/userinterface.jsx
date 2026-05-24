import React, {Component} from 'react';
import { AuthContext } from './authContext';

export default class UserInterface extends Component {
    static contextType = AuthContext;

    render() {
        const { currentUser, playerGold, troops, buyTroop } = this.context;
        if (!currentUser) { return null; }

        const gold = parseInt(playerGold) || 0;

        const panelStyle = {
            position: 'absolute',
            zIndex: 100,
            top: '235px',
            right: '22px',
            width: '260px',
            padding: '14px',
            border: '2px solid black',
            backgroundColor: '#e6e6e6',
            display: 'grid',
            gap: '12px'
        };
        const rowStyle = {
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            gap: '10px'
        };
        const buttonStyle = (cost) => ({
            padding: '6px 10px',
            border: '1px solid black',
            backgroundColor: '#fff',
            fontWeight: 'bold',
            cursor: gold >= cost ? 'pointer' : 'not-allowed',
            opacity: gold >= cost ? 1 : 0.5,
        });

        return (
            <div style={panelStyle}>
                <div>Bonjour {currentUser}</div>
                <div>💰 Or: {gold/1000}K</div>
                <div style={{ borderTop: '1px solid #999', paddingTop: '8px', display: 'grid', gap: '8px' }}>
                    <div style={{ fontWeight: 'bold' }}>Troupes</div>
                    <div style={rowStyle}>
                        <span>🪖 Soldats: {troops.soldiers}</span>
                        <button style={buttonStyle(10)} onClick={() => buyTroop('soldiers')} disabled={gold < 10}>
                            Acheter 1K or
                        </button>
                    </div>
                    <div style={rowStyle}>
                        <span>🦖 Tanks: {troops.tanks}</span>
                        <button style={buttonStyle(20)} onClick={() => buyTroop('tanks')} disabled={gold < 20}>
                            Acheter 2K or
                        </button>
                    </div>
                    <div style={rowStyle}>
                        <span>✈️ Avions: {troops.planes}</span>
                        <button style={buttonStyle(30)} onClick={() => buyTroop('planes')} disabled={gold < 30}>
                            Acheter 3K or
                        </button>
                    </div>
                </div>
            </div>
        );
    }
}