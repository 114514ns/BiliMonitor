import React, {useEffect, useRef, useState} from 'react';

function Danmaku(props) {

    const [logs, setLogs] = useState([]);
    const logsEndRef = useRef(null);

    // 模拟获取日志数据
    const fetchLogs = () => {
        setLogs(prevLogs => [...prevLogs, `New log at ${new Date().toLocaleTimeString()}`]);
    };

    // 每隔5秒刷新一次日志
    useEffect(() => {
        const interval = setInterval(fetchLogs, 500);

        return () => clearInterval(interval);
    }, []);

    useEffect(() => {
        if (logsEndRef.current) {
            logsEndRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [logs]);

    return (
        <div style={{  overflowY: 'auto', border: '1px solid #ddd' }}>
            <ul style={{ padding: 0 }}>
                {logs.map((log, index) => (
                    <li key={index} style={{ padding: '5px 0', borderBottom: '1px solid #eee' }}>
                        {log}
                    </li>
                ))}
            </ul>
            <div ref={logsEndRef} />
        </div>
    );
}

export default Danmaku;