import React, {Component} from 'react';
import { AuthContext } from './authContext';

export default class UserInterface extends Component {
    static contextType = AuthContext;

    render() {
        const { currentUser, playerGold, troops, buyTroop } = this.context;

        if (!currentUser) { return null; }

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

        const buttonStyle = {
            padding: '6px 10px',
            border: '1px solid black',
            backgroundColor: '#fff',
            cursor: 'pointer',
            fontWeight: 'bold'
        };

        return (
            <div style={panelStyle}>
                <div>
                    Bonjour {currentUser}
                </div>
                <div>💰 Or: {playerGold}</div>
                <div style={{ borderTop: '1px solid #999', paddingTop: '8px', display: 'grid', gap: '8px' }}>
                    <div style={{ fontWeight: 'bold' }}>Troupes</div>
                    <div style={rowStyle}>
                        <span>🪖 Soldats: {troops.soldiers}</span>
                        <button style={buttonStyle} onClick={() => buyTroop('soldiers')} disabled={playerGold < 1}>
                            Acheter 10 or
                        </button>
                    </div>
                    <div style={rowStyle}>
                        <span>🦖 Tanks: {troops.tanks}</span>
                        <button style={buttonStyle} onClick={() => buyTroop('tanks')} disabled={playerGold < 2}>
                            Acheter 20 or
                        </button>
                    </div>
                    <div style={rowStyle}>
                        <span>✈️ Avions: {troops.planes}</span>
                        <button style={buttonStyle} onClick={() => buyTroop('planes')} disabled={playerGold < 3}>
                            Acheter 30 or
                        </button>
                    </div>
                </div>
            </div>
        );
    }

}