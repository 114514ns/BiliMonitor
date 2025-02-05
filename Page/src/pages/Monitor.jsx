import classes from "./Monitor.module.css";
import {useEffect, useState} from "react";
import axios from "axios"
import {Card} from "antd";
import LiveCard from "../components/LiveCard.jsx";

const Monitor = () => {
    const [monitor,setMonitor] = useState([])

    useEffect(() => {
        axios.get("http://localhost:8080/monitor").then(res => {
            //setMonitor(res.data.lives)
            res.data.lives.sort((a, b) => {
                if(a.Live === b.Live) return 0; // 当前后两个元素的isOnline相同时, 不改变顺序
                return a.Live ? -1 : 1; // 如果a的isOnline为true, 那么让a排在b前面, 否则排在后面
            });
            setMonitor(res.data.lives)
        })
    }, [])

    return (
        <div className={classes.container}>
            {monitor.map((item, index) => {
                return <LiveCard key={index} liveData={item} />
            })}
        </div>
    );
};

export default Monitor;