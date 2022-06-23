import React, {FC} from "react";
import {Link} from "react-router-dom";

const Footer: FC = () => {
	return <footer id="footer">
		<div id="footer-nav">
			<div className="wrap">
				<ul>
					<li>
						<a href="//github.com/udovin/goquiz">Repository</a>
					</li>
					<li>Language: <Link to="/language">English</Link></li>
				</ul>
			</div>
		</div>
		<div id="footer-copy">
			<a href="//github.com/udovin">Ivan Udovin</a> &copy; 2022
		</div>
	</footer>;
};

export default Footer;
