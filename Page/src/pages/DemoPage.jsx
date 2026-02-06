import React from 'react';
import BVPlayer from "../components/BVPlayer";
import HeatMap, {HeatContent} from "../components/HeatChart";


function DemoPage(props) {
    return (
        <div>
           {/* <BVPlayer bv={'BV1Md2EBDEtZ'}/>*/}
            <HeatContent uid={3691006776052082}/>
        </div>
    );
}

export default DemoPage;