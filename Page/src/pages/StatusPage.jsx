import React, {useEffect} from 'react';
import axios from "axios";

function StatusPage(props) {

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
                    className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">每分钟弹幕<br/><span
                    className="font-semibold">{overview.Message1}</span>
                </div>
                <div
                    className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">每5分钟弹幕<br/>
                    <span className="font-semibold">{overview.Message5}</span>
                </div>
                <div
                    className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">每小时弹幕<br/>
                    <span
                        className="font-semibold">{overview.MessageHour}</span>
                </div>

            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm mt-4">
                <div
                    className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">24小时弹幕<br/>
                    <span
                        className="font-semibold">{overview.MessageDaily}</span>
                </div>
                <div
                    className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">总弹幕<br/>
                    <span
                        className="font-semibold">{overview.TotalMessages}</span>
                </div>
            </div>
        </div>
    );
}

export default StatusPage;