import React, { Component } from 'react';

export default class Login extends Component {

    state = {
        logged: false
    };

    componentDidMount() {
        fetch('/me', { credentials: 'include' })
            .then(res => {
                if (!res.ok) throw new Error();
                return res.json();
            })
            .then(() => this.setState({ logged: true }))
            .catch(() => this.setState({ logged: false }));
    }

    render() {
        return this.state.logged
            ? <a href="/logout">Se déconnecter</a>
            : <a href="/login.html">Se connecter</a>;
    }
}