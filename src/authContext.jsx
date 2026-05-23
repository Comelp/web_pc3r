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
                    troops: { soldiers: 0, tanks: 0, planes: 0 },
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

    buyTroop = (troopType) => { // Todo: send to server and refresh gold/troops from server response
        const costs = {
            soldiers: 10,
            tanks: 20,
            planes: 30
        };

        const cost = costs[troopType];
        if (!cost || !this.state.currentUser) {
            return;
        }

        this.setState(prevState => {
            if (prevState.playerGold < cost) {
                return null;
            }

            return {
                playerGold: prevState.playerGold - cost,
                troops: {
                    ...prevState.troops,
                    [troopType]: prevState.troops[troopType] + 1
                }
            };
        });
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