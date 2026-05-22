import React, { useState, useEffect } from 'react';

export default function Timer() {
    const [phase, setPhase] = useState('');
    const [timeLeft, setTimeLeft] = useState('00:00');
    const [shouldFetch, setShouldFetch] = useState(true);

    useEffect(() => {
        if (shouldFetch) {
            fetch('/getPhase')
                .then((res) => res.text())
                .then((text) => {
                    try {
                        const data = JSON.parse(text);
                        setPhase(data.phase);
                    } catch (err) {
                        console.error('Invalid JSON from /getPhase:', text, err);
                    }
                })
                .catch((err) => console.error('Error fetching phase:', err));
            setShouldFetch(false);
        }
    }, [shouldFetch]);

    useEffect(() => {
        const interval = setInterval(() => {
            const now = new Date();
            const secondsInCycle = now.getSeconds() + now.getMinutes() * 60;
            const remaining = 600 - (secondsInCycle % 600);

            if (remaining === 600) {
                setShouldFetch(true);
            }

            const minutes = Math.floor(remaining / 60);
            const seconds = remaining % 60;
            setTimeLeft(
                `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
            );
        }, 1000);

        return () => clearInterval(interval);
    }, []);

    return (
        <div>
            <p style={{ fontSize: '2em' }}>Phase actuelle: {phase}</p>
            <p style={{ fontSize: '2em' }}>Prochaine phase dans: {timeLeft}</p>
        </div>
    );
}