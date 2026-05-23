import { AuthContext } from './authContext';
import React, { Component } from 'react';

export default class DeployForm extends Component {
    static contextType = AuthContext;

    state = { troop: 'soldiers', count: 1, error: null };

    submit() {
        const { troop, count } = this.state;
        const { mode, countryName, onDeployed } = this.props;

        fetch(`/deployTroop?country=${countryName}&mode=${mode}&troop=${troop}&count=${count}`, {
            method: 'POST',
            credentials: 'include'
        })
        .then(async res => {
            const text = await res.text();
            if (!res.ok) throw new Error(text);
            this.context.refreshAuth();
            onDeployed();
        })
        .catch(err => this.setState({ error: err.message }));
    }

    render() {
        const { troop, count, error } = this.state;
        const { mode, onBack } = this.props;
        const { troops } = this.context;
        const available = troops[troop] ?? 0;

        return (
            <>
                <p><strong>{mode === 'defending' ? 'Déployer la Défense' : "Déployer l'Attaque"}</strong></p>
                <div>
                    <label>Type : </label>
                    <select value={troop} onChange={e => this.setState({ troop: e.target.value, error: null })}>
                        <option value="soldiers">🪖 Soldats ({troops.soldiers})</option>
                        <option value="tanks">🦖 Tanks ({troops.tanks})</option>
                        <option value="planes">✈️ Avions ({troops.planes})</option>
                    </select>
                </div>
                <div style={{ marginTop: '6px' }}>
                    <label>Quantité (max {available}) : </label>
                    <input
                        type="number" min="1" max={available} value={count}
                        onChange={e => this.setState({ count: parseInt(e.target.value) || 1, error: null })}
                        style={{ width: '50px' }}
                    />
                </div>
                {error && <p style={{ color: 'red', fontSize: '12px' }}>{error}</p>}
                <div style={{ marginTop: '8px', display: 'flex', gap: '8px' }}>
                    <button onClick={() => onBack()}>← Retour</button>
                    <button
                        onClick={() => this.submit()}
                        disabled={count < 1 || count > available}
                    >
                        Confirmer
                    </button>
                </div>
            </>
        );
    }
}