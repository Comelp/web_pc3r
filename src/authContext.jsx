import React, { createContext, Component } from 'react';

export const AuthContext = createContext({
    currentUser: null,
    playerGold: 0,
    troops: { soldiers: 0, tanks: 0, planes: 0 },
    refreshAuth: () => {},
    buyTroop: () => {}
});

export class AuthProvider extends Component {
    state = {
        currentUser: null,
        playerGold: 0,
        troops: { soldiers: 0, tanks: 0, planes: 0 },
        isLoading: true
    };

    componentDidMount() {
        this.refreshAuth();
    }

    refreshAuth = () => {
        fetch('/me', { credentials: 'include' })
            .then(res => {
                if (!res.ok) {
                    this.setState({
                        currentUser: null,
                        playerGold: 0,
                        troops: { soldiers: 0, tanks: 0, planes: 0 },
                        isLoading: false
                    });
                    return null;
                }
                return res.json();
            })
            .then(data => {
                if (!data) return;
                this.setState({
                    currentUser: data.username ?? null,
                    playerGold: data.gold ?? 0,
                    troops: {
                        soldiers: data.troops?.soldiers ?? 0,
                        tanks:    data.troops?.tanks    ?? 0,
                        planes:   data.troops?.planes   ?? 0,
                    },
                    isLoading: false
                });
            })
            .catch(() => this.setState({
                currentUser: null,
                playerGold: 0,
                troops: { soldiers: 0, tanks: 0, planes: 0 },
                isLoading: false
            }));
    };

    buyTroop = (troopType) => {
        if (!this.state.currentUser) return;

        fetch(`/buyTroop?troop=${troopType}`, {
            method: 'POST',
            credentials: 'include'
        })
        .then(async res => {
            const text = await res.text();
            if (!res.ok) throw new Error(text);
            return JSON.parse(text);
        })
        .then(data => {
            this.setState({
                playerGold: data.gold_remaining,
                troops: {
                    soldiers: data.troops.soldiers ?? 0,
                    tanks:    data.troops.tanks    ?? 0,
                    planes:   data.troops.planes   ?? 0,
                }
            });
        })
        .catch(err => alert(err.message));
    };

    render() {
        const value = {
            currentUser: this.state.currentUser,
            playerGold: this.state.playerGold,
            troops: this.state.troops,
            isLoading: this.state.isLoading,
            refreshAuth: this.refreshAuth,
            buyTroop: this.buyTroop
        };

        return (
            <AuthContext.Provider value={value}>
                {this.props.children}
            </AuthContext.Provider>
        );
    }
}