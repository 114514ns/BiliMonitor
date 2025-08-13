import React, {useEffect} from 'react';
import axios from "axios";

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
    useEffect(() => {
        var intervalId = setInterval(() => {
            axios.get('/api/status').then(res => {
                setOverview(res.data);
            })
        },500)
        return () => clearInterval(intervalId);
    },[])
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
        </div>
    );
}

export default StatusPage;