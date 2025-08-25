import React, {useEffect} from 'react';
import axios from "axios";
import {LineChart} from "@mui/x-charts/LineChart";

function StatusPage(props) {

    function formatNumber(num, {
        decimals = 1,
        suffix = '',
    } = {}) {
        if (num === null || num === undefined || isNaN(num)) return '0';

        const units = ['', 'K', 'M', 'G', 'T', 'P'];
        const tier = Math.floor(Math.log10(Math.abs(num)) / 3);

        if (tier === 0) return num + suffix;

        const unit = units[tier];
        const scaled = num / Math.pow(10, tier * 3);
        return scaled.toFixed(decimals) + unit + suffix;
    }


    const [overview, setOverview] = React.useState({});

    const [msg,setMsg] = React.useState([]);
    useEffect(() => {
        var intervalId = setInterval(() => {
            axios.get('/api/status').then(res => {
                setOverview(res.data);
            })
        },500)
        return () => clearInterval(intervalId);
    },[])
    useEffect(() => {
        axios.get('/api/chart/msg').then(res => {
            setMsg(res.data.data);
        })
    },[])

    const contentMap = ['24小时弹幕','月弹幕','6月弹幕']
    return (

        <div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm">
                <div
                    className="rounded-xl bg-pink-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">每分钟弹幕<br/><span
                    className="font-semibold">{overview.Message1}</span>
                </div>
                <div
                    className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">每5分钟弹幕<br/>
                    <span className="font-semibold">{overview.Message5}</span>
                </div>
                <div
                    className="rounded-xl bg-green-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">每小时弹幕<br/>
                    <span
                        className="font-semibold">{overview.MessageHour}</span>
                </div>

            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm mt-4">
                <div
                    className="rounded-xl bg-orange-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">24小时弹幕<br/>
                    <span
                        className="font-semibold">{overview.MessageDaily}</span>
                </div>
                <div
                    className="rounded-xl bg-blue-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">总弹幕<br/>
                    <span
                        className="font-semibold">{overview.TotalMessages}</span>
                </div>
                <div
                    className="rounded-xl bg-blue-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">启动时间<br/>
                    <span
                        className="font-semibold">{overview.LaunchedAt}</span>
                </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm mt-4">
                <div
                    className="rounded-xl bg-orange-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">HTTP请求书<br/>
                    <span
                        className="font-semibold">{formatNumber(overview.Requests)}</span>
                </div>
                <div
                    className="rounded-xl bg-yellow-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">HTTP字节数<br/>
                    <span
                        className="font-semibold">{formatNumber(overview.HTTPBytes)}</span>
                </div>
                <div
                    className="rounded-xl bg-purple-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">Websocket字节数<br/>
                    <span
                        className="font-semibold">{formatNumber(overview.WSBytes)}</span>
                </div>
            </div>
            {msg.length > 0 &&
                msg.map((msg, i) => (
                    <LineChart

                        xAxis={[{
                            data: msg.map((item, index) => (new Date(item.Time).getTime())),
                            label: contentMap[i],
                            valueFormatter: (timestamp) => {
                                const date = new Date(timestamp)
                                if (i !== 0) {
                                    return date.toLocaleDateString()
                                }
                                return date.toLocaleTimeString()
                            }
                        }]}
                        series={[
                            {
                                area:true,
                                data: msg.map((item, index) => (item.Count))
                            },
                        ]}
                        height={300}
                    />
                ))
            }
        </div>
    );
}

export default StatusPage;