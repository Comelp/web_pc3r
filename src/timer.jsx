import React, { useState, useEffect } from 'react';

export default function Timer() {
    const [phase, setPhase] = useState('');
    const [timeLeft, setTimeLeft] = useState('00:00');

    useEffect(() => {
        const fetchPhase = () => {
            fetch('/getPhase')
                .then(res => res.json())
                .then(data => setPhase(data.phase))
                .catch(err => console.error(err));
        };

        fetchPhase();
        const intervalPhase = setInterval(fetchPhase, 5000);

        const intervalTimer = setInterval(() => {
            const now = new Date();
            const secondsInCycle = now.getSeconds() + now.getMinutes() * 60;
            const remaining = 600 - (secondsInCycle % 600);

            const minutes = Math.floor(remaining / 60);
            const seconds = remaining % 60;

            setTimeLeft(
                `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
            );
        }, 1000);

        return () => {
            clearInterval(intervalPhase);
            clearInterval(intervalTimer);
        };
    }, []);

    return (
        <div>
            <p style={{ fontSize: '2em' }}>Phase actuelle: {phase}</p>
            <p style={{ fontSize: '2em' }}>Prochaine phase dans: {timeLeft}</p>
        </div>
    );
}